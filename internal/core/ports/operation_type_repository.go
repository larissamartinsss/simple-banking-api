package ports

import (
	"context"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
)

// OperationTypeRepository defines the interface for operation type data operations
type OperationTypeRepository interface {
	FindByID(ctx context.Context, id int) (*domain.OperationType, error)
	GetAll(ctx context.Context) ([]*domain.OperationType, error)
	// Seed initializes the database with the predefined operation types - This should be called during application startup
	Seed(ctx context.Context) error
}
