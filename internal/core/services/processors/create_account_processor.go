package processors

import (
	"context"
	"errors"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

type CreateAccountProcessor struct {
	accountRepo ports.AccountRepository
}

func NewCreateAccountProcessor(accountRepo ports.AccountRepository) *CreateAccountProcessor {
	return &CreateAccountProcessor{
		accountRepo: accountRepo,
	}
}

func (p *CreateAccountProcessor) Process(ctx context.Context, req domain.CreateAccountRequest) (*domain.CreateAccountResponse, error) {
	account := &domain.Account{DocumentNumber: req.DocumentNumber}

	// Check if account with this document number already exists
	existing, err := p.accountRepo.FindByDocumentNumber(ctx, req.DocumentNumber)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		return nil, errors.New("account with this document number already exists")
	}

	// Create the account in repository
	createdAccount, err := p.accountRepo.Create(ctx, account)
	if err != nil {
		return nil, err
	}

	return &domain.CreateAccountResponse{
		Account: createdAccount,
	}, nil
}
