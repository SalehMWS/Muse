package publishing

import (
	"context"
	"encoding/json"
	"fmt"

	pubapp "github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/worker/domain"
)

type PublishHandler struct {
	publish *pubapp.PublishUseCase
}

func NewPublishHandler(publish *pubapp.PublishUseCase) *PublishHandler {
	return &PublishHandler{publish: publish}
}

func (h *PublishHandler) Handle(ctx context.Context, job domain.Job) error {
	var payload domain.PublishPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("decode publish payload: %w", err)
	}

	_, err := h.publish.Execute(ctx, pubapp.PublishInput{
		UserID:             payload.UserID,
		ContentID:          payload.ContentID,
		InstagramAccountID: payload.InstagramAccountID,
		MediaType:          payload.MediaType,
	})
	return err
}
