package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SalehMWS/Muse/internal/worker/domain"
)

func newTestPool(broker Broker, handler Handler) *Pool {
	dispatcher := NewDispatcher()
	dispatcher.Register(domain.TypeInstagramPublish, handler)
	return NewPool(broker, dispatcher, nil, PoolOptions{Workers: 1, Block: time.Millisecond})
}

func job(attempt, maxAttempts int) domain.Job {
	return domain.Job{ID: "job-1", Type: domain.TypeInstagramPublish, Attempt: attempt, MaxAttempts: maxAttempts}
}

func TestPool_SuccessAcks(t *testing.T) {
	broker := &fakeBroker{}
	pool := newTestPool(broker, HandlerFunc(func(context.Context, domain.Job) error { return nil }))

	pool.handle(context.Background(), Delivery{Job: job(0, 3), Reference: "m1"})

	if len(broker.acked) != 1 || broker.acked[0] != "m1" {
		t.Fatalf("acked = %v, want [m1]", broker.acked)
	}
	if s := pool.Stats(); s.Succeeded != 1 || s.Processed != 1 || s.Failed != 0 {
		t.Fatalf("stats = %+v, unexpected", s)
	}
}

func TestPool_FailureRetries(t *testing.T) {
	broker := &fakeBroker{}
	pool := newTestPool(broker, HandlerFunc(func(context.Context, domain.Job) error { return errors.New("boom") }))

	pool.handle(context.Background(), Delivery{Job: job(0, 3), Reference: "m1"})

	if len(broker.enqueued) != 1 || broker.enqueued[0].Attempt != 1 {
		t.Fatalf("enqueued = %+v, want one requeued job at attempt 1", broker.enqueued)
	}
	if len(broker.acked) != 1 || len(broker.dead) != 0 {
		t.Fatalf("acked=%v dead=%v, want acked original and no dead-letter", broker.acked, broker.dead)
	}
	if s := pool.Stats(); s.Retried != 1 || s.Failed != 1 {
		t.Fatalf("stats = %+v, want retried=1 failed=1", s)
	}
}

func TestPool_ExhaustedDeadLetters(t *testing.T) {
	broker := &fakeBroker{}
	pool := newTestPool(broker, HandlerFunc(func(context.Context, domain.Job) error { return errors.New("boom") }))

	pool.handle(context.Background(), Delivery{Job: job(2, 3), Reference: "m1"})

	if len(broker.dead) != 1 || len(broker.enqueued) != 0 {
		t.Fatalf("dead=%v enqueued=%v, want one dead-letter and no requeue", broker.dead, broker.enqueued)
	}
	if len(broker.acked) != 1 {
		t.Fatalf("acked = %v, want original acked after dead-letter", broker.acked)
	}
	if s := pool.Stats(); s.DeadLettered != 1 {
		t.Fatalf("stats = %+v, want deadLettered=1", s)
	}
}

func TestPool_UnknownTypeIsFailure(t *testing.T) {
	broker := &fakeBroker{}
	dispatcher := NewDispatcher()
	pool := NewPool(broker, dispatcher, nil, PoolOptions{Workers: 1, Block: time.Millisecond})

	pool.handle(context.Background(), Delivery{Job: domain.Job{ID: "x", Type: "unknown", Attempt: 2, MaxAttempts: 3}, Reference: "m1"})

	if len(broker.dead) != 1 {
		t.Fatalf("unknown job type should dead-letter after exhaustion; dead=%v", broker.dead)
	}
}
