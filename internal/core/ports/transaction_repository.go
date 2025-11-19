package ports

import (
	"context"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
)

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error)
	FindByID(ctx context.Context, id int) (*domain.Transaction, error)
	FindByAccountID(ctx context.Context, accountID int) ([]*domain.Transaction, error)
	GetAll(ctx context.Context) ([]*domain.Transaction, error)
}
