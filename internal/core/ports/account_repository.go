package ports

import (
	"context"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
)

// AccountRepository defines the interface for account data operations
// This is a port in hexagonal architecture - it defines WHAT we need without HOW
type AccountRepository interface {
	Create(ctx context.Context, account *domain.Account) (*domain.Account, error)
	FindByID(ctx context.Context, id int) (*domain.Account, error)
	FindByDocumentNumber(ctx context.Context, documentNumber string) (*domain.Account, error)
	GetAll(ctx context.Context) ([]*domain.Account, error)
}
