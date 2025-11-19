package processors

import (
	"context"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
)

type CreateAccountProcessorInterface interface {
	Process(ctx context.Context, req domain.CreateAccountRequest) (*domain.CreateAccountResponse, error)
}

type GetAccountProcessorInterface interface {
	Process(ctx context.Context, req domain.GetAccountRequest) (*domain.GetAccountResponse, error)
}

type CreateTransactionProcessorInterface interface {
	Process(ctx context.Context, req domain.CreateTransactionRequest) (*domain.CreateTransactionResponse, error)
}
