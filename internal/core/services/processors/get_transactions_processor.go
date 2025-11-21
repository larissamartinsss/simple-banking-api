package processors

import (
	"context"
	"fmt"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

// GetTransactionsProcessor handles the business logic for getting transactions
type GetTransactionsProcessor struct {
	transactionRepo ports.TransactionRepository
	accountRepo     ports.AccountRepository
}

// NewGetTransactionsProcessor creates a new GetTransactionsProcessor
func NewGetTransactionsProcessor(transactionRepo ports.TransactionRepository, accountRepo ports.AccountRepository) *GetTransactionsProcessor {
	return &GetTransactionsProcessor{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
	}
}

func (p *GetTransactionsProcessor) Process(ctx context.Context, req domain.GetTransactionsRequest) (*domain.GetTransactionsResponse, error) {

	// Validate account exists
	account, err := p.accountRepo.FindByID(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account with id %d not found", req.AccountID)
	}

	transactions, total, err := p.transactionRepo.FindByAccountIDPaginated(ctx, req.AccountID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}

	p.validatePagination(req)

	// Calculate number of pages
	pages := (total + req.Limit - 1) / req.Limit
	if pages < 1 {
		pages = 1
	}

	// Build response
	return &domain.GetTransactionsResponse{
		Transactions: transactions,
		Pagination: domain.PaginationMetadata{
			Total:  total,
			Limit:  req.Limit,
			Offset: req.Offset,
			Pages:  pages,
		},
	}, nil
}

func (p *GetTransactionsProcessor) validatePagination(req domain.GetTransactionsRequest) {
	// Validate pagination parameters
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50 // Default
	}
	if req.Offset < 0 {
		req.Offset = 0 // Default
	}

}
