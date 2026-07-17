package publishing

import (
	"context"

	pubapp "github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/scheduler/application"
)

type Publisher struct {
	publish *pubapp.PublishUseCase
}

func NewPublisher(publish *pubapp.PublishUseCase) *Publisher {
	return &Publisher{publish: publish}
}

func (p *Publisher) Publish(ctx context.Context, cmd application.PublishCommand) error {
	_, err := p.publish.Execute(ctx, pubapp.PublishInput{
		UserID:             cmd.UserID,
		ContentID:          cmd.ContentID,
		InstagramAccountID: cmd.InstagramAccountID,
		MediaType:          cmd.MediaType,
	})
	return err
}
