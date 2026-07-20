package application

import "errors"

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrEmptyContent     = errors.New("content is required")
	ErrEmptyQuery       = errors.New("query is required")
	ErrEmbedding        = errors.New("embedding failed")
	ErrVectorStore      = errors.New("vector store operation failed")
)
