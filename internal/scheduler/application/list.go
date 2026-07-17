package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

type ListSchedulesUseCase struct {
	repo ScheduleRepository
}

func NewListSchedulesUseCase(repo ScheduleRepository) *ListSchedulesUseCase {
	return &ListSchedulesUseCase{repo: repo}
}

func (uc *ListSchedulesUseCase) Execute(ctx context.Context, userID, contentID uuid.UUID) ([]domain.Schedule, error) {
	return uc.repo.ListByContentForUser(ctx, userID, contentID)
}

type CancelScheduleUseCase struct {
	repo ScheduleRepository
}

func NewCancelScheduleUseCase(repo ScheduleRepository) *CancelScheduleUseCase {
	return &CancelScheduleUseCase{repo: repo}
}

func (uc *CancelScheduleUseCase) Execute(ctx context.Context, userID, scheduleID uuid.UUID) error {
	if _, err := uc.repo.FindByIDForUser(ctx, scheduleID, userID); err != nil {
		return err
	}
	return uc.repo.Cancel(ctx, scheduleID, userID)
}
