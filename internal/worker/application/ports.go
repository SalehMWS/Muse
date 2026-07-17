package application

import (
	"context"
	"time"

	"github.com/SalehMWS/Muse/internal/worker/domain"
)

type Enqueuer interface {
	Enqueue(ctx context.Context, job domain.Job) error
}

type Delivery struct {
	Job       domain.Job
	Reference string
}

type Broker interface {
	Enqueue(ctx context.Context, job domain.Job) error
	Read(ctx context.Context, count int, block time.Duration) ([]Delivery, error)
	Ack(ctx context.Context, delivery Delivery) error
	DeadLetter(ctx context.Context, job domain.Job, reason string) error
}

type Handler interface {
	Handle(ctx context.Context, job domain.Job) error
}

type HandlerFunc func(ctx context.Context, job domain.Job) error

func (f HandlerFunc) Handle(ctx context.Context, job domain.Job) error {
	return f(ctx, job)
}
