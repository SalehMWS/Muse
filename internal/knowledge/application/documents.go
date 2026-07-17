package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/domain"
)

type ListDocumentsUseCase struct {
	repo DocumentRepository
}

func NewListDocumentsUseCase(repo DocumentRepository) *ListDocumentsUseCase {
	return &ListDocumentsUseCase{repo: repo}
}

func (uc *ListDocumentsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]domain.Document, error) {
	return uc.repo.ListByUser(ctx, userID)
}

type DeleteDocumentUseCase struct {
	repo  DocumentRepository
	store VectorStore
}

func NewDeleteDocumentUseCase(repo DocumentRepository, store VectorStore) *DeleteDocumentUseCase {
	return &DeleteDocumentUseCase{repo: repo, store: store}
}

func (uc *DeleteDocumentUseCase) Execute(ctx context.Context, userID, documentID uuid.UUID) error {
	if _, err := uc.repo.FindByIDForUser(ctx, documentID, userID); err != nil {
		return err
	}
	if err := uc.store.DeleteByDocument(ctx, userID, documentID); err != nil {
		return err
	}
	return uc.repo.DeleteForUser(ctx, documentID, userID)
}
