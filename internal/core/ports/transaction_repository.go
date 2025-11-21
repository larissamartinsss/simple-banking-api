package ports

import (
	"context"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
)

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error)
	FindByID(ctx context.Context, id int64) (*domain.Transaction, error)
	FindByAccountID(ctx context.Context, accountID int64) ([]*domain.Transaction, error)
	GetAll(ctx context.Context) ([]*domain.Transaction, error)
	FindByAccountIDPaginated(ctx context.Context, accountID int64, limit int64, offset int64) ([]*domain.Transaction, int64, error)
}
