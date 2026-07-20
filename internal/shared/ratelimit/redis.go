package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "novaflow:ratelimit:"

var slidingWindow = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

redis.call('ZREMRANGEBYSCORE', key, 0, now - window)
local used = redis.call('ZCARD', key)

if used >= limit then
  local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
  local retry = window
  if oldest[2] then
    retry = tonumber(oldest[2]) + window - now
  end
  if retry < 1 then
    retry = 1
  end
  return {0, 0, retry}
end

redis.call('ZADD', key, now, member)
redis.call('PEXPIRE', key, math.ceil(window / 1000))
return {1, limit - used - 1, 0}
`)

type RedisLimiter struct {
	client redis.Scripter
}

func NewRedisLimiter(client redis.Scripter) *RedisLimiter {
	return &RedisLimiter{client: client}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (Decision, error) {
	if limit <= 0 || window <= 0 {
		return Decision{Allowed: true, Limit: limit, Remaining: limit}, nil
	}

	values, err := slidingWindow.Run(
		ctx,
		l.client,
		[]string{keyPrefix + key},
		time.Now().UnixMicro(),
		window.Microseconds(),
		limit,
		uuid.NewString(),
	).Slice()
	if err != nil {
		return Decision{}, fmt.Errorf("ratelimit: evaluate window: %w", err)
	}

	if len(values) != 3 {
		return Decision{}, fmt.Errorf("ratelimit: unexpected script result length %d", len(values))
	}

	allowed, err := toInt64(values[0])
	if err != nil {
		return Decision{}, err
	}

	remaining, err := toInt64(values[1])
	if err != nil {
		return Decision{}, err
	}

	retryAfter, err := toInt64(values[2])
	if err != nil {
		return Decision{}, err
	}

	return Decision{
		Allowed:    allowed == 1,
		Limit:      limit,
		Remaining:  int(remaining),
		RetryAfter: time.Duration(retryAfter) * time.Microsecond,
	}, nil
}

func toInt64(value any) (int64, error) {
	parsed, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("ratelimit: unexpected script result type %T", value)
	}
	return parsed, nil
}
