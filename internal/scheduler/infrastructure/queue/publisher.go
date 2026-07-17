package queue

import (
	"context"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	workerapp "github.com/SalehMWS/Muse/internal/worker/application"
	workerdomain "github.com/SalehMWS/Muse/internal/worker/domain"
)

const defaultMaxAttempts = 3

type Publisher struct {
	enqueuer    workerapp.Enqueuer
	maxAttempts int
}

func NewPublisher(enqueuer workerapp.Enqueuer, maxAttempts int) *Publisher {
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}
	return &Publisher{enqueuer: enqueuer, maxAttempts: maxAttempts}
}

func (p *Publisher) Publish(ctx context.Context, cmd application.PublishCommand) error {
	job, err := workerdomain.NewJob(workerdomain.TypeInstagramPublish, workerdomain.PublishPayload{
		UserID:             cmd.UserID,
		ContentID:          cmd.ContentID,
		InstagramAccountID: cmd.InstagramAccountID,
		MediaType:          cmd.MediaType,
	}, p.maxAttempts)
	if err != nil {
		return err
	}
	return p.enqueuer.Enqueue(ctx, job)
}
