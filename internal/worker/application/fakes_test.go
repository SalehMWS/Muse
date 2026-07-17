package application

import (
	"context"
	"time"

	"github.com/SalehMWS/Muse/internal/worker/domain"
)

type fakeBroker struct {
	enqueued []domain.Job
	acked    []string
	dead     []domain.Job
}

func (b *fakeBroker) Enqueue(_ context.Context, job domain.Job) error {
	b.enqueued = append(b.enqueued, job)
	return nil
}

func (b *fakeBroker) Read(context.Context, int, time.Duration) ([]Delivery, error) {
	return nil, nil
}

func (b *fakeBroker) Ack(_ context.Context, delivery Delivery) error {
	b.acked = append(b.acked, delivery.Reference)
	return nil
}

func (b *fakeBroker) DeadLetter(_ context.Context, job domain.Job, _ string) error {
	b.dead = append(b.dead, job)
	return nil
}
