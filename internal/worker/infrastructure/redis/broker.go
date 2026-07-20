package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/SalehMWS/Muse/internal/shared/tracing"
	"github.com/SalehMWS/Muse/internal/worker/application"
	"github.com/SalehMWS/Muse/internal/worker/domain"
)

const jobField = "job"

type Broker struct {
	client   *goredis.Client
	stream   string
	dlq      string
	group    string
	consumer string
}

var _ application.Broker = (*Broker)(nil)

func NewBroker(client *goredis.Client, stream, group, consumer string) *Broker {
	return &Broker{
		client:   client,
		stream:   stream,
		dlq:      stream + ":dead",
		group:    group,
		consumer: consumer,
	}
}

func (b *Broker) EnsureGroup(ctx context.Context) error {
	err := b.client.XGroupCreateMkStream(ctx, b.stream, b.group, "$").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}
	return nil
}

func (b *Broker) Enqueue(ctx context.Context, job domain.Job) error {
	if job.TraceID == "" {
		ids := tracing.FromContext(ctx)
		job = job.WithTrace(ids.TraceID, ids.CorrelationID)
	}

	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return b.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: b.stream,
		Values: map[string]any{jobField: data},
	}).Err()
}

func (b *Broker) Depth(ctx context.Context) (int64, error) {
	return b.client.XLen(ctx, b.stream).Result()
}

func (b *Broker) DeadLetterDepth(ctx context.Context) (int64, error) {
	return b.client.XLen(ctx, b.dlq).Result()
}

func (b *Broker) Stream() string {
	return b.stream
}

func (b *Broker) DeadLetterStream() string {
	return b.dlq
}

func (b *Broker) DeadLetter(ctx context.Context, job domain.Job, reason string) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return b.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: b.dlq,
		Values: map[string]any{jobField: data, "reason": reason},
	}).Err()
}

func (b *Broker) Read(ctx context.Context, count int, block time.Duration) ([]application.Delivery, error) {
	streams, err := b.client.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    b.group,
		Consumer: b.consumer,
		Streams:  []string{b.stream, ">"},
		Count:    int64(count),
		Block:    block,
	}).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var deliveries []application.Delivery
	for _, stream := range streams {
		for _, message := range stream.Messages {
			raw, _ := message.Values[jobField].(string)
			var job domain.Job
			if err := json.Unmarshal([]byte(raw), &job); err != nil {
				_ = b.client.XAck(ctx, b.stream, b.group, message.ID).Err()
				continue
			}
			deliveries = append(deliveries, application.Delivery{Job: job, Reference: message.ID})
		}
	}
	return deliveries, nil
}

func (b *Broker) Ack(ctx context.Context, delivery application.Delivery) error {
	return b.client.XAck(ctx, b.stream, b.group, delivery.Reference).Err()
}
