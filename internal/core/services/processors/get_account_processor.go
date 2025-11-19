package processors

import (
	"context"
	"errors"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

type GetAccountProcessor struct {
	accountRepo ports.AccountRepository
}

func NewGetAccountProcessor(accountRepo ports.AccountRepository) *GetAccountProcessor {
	return &GetAccountProcessor{
		accountRepo: accountRepo,
	}
}

func (p *GetAccountProcessor) Process(ctx context.Context, req domain.GetAccountRequest) (*domain.GetAccountResponse, error) {
	// Get account from repository
	account, err := p.accountRepo.FindByID(ctx, req.AccountID)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return nil, errors.New("account not found")
	}

	return &domain.GetAccountResponse{
		Account: account,
	}, nil
}
