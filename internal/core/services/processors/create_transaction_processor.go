package processors

import (
	"context"
	"fmt"
	"time"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

// CreateTransactionProcessor handles the business logic for creating a transaction
type CreateTransactionProcessor struct {
	transactionRepo   ports.TransactionRepository
	accountRepo       ports.AccountRepository
	operationTypeRepo ports.OperationTypeRepository
}

// NewCreateTransactionProcessor creates a new CreateTransactionProcessor
func NewCreateTransactionProcessor(transactionRepo ports.TransactionRepository, accountRepo ports.AccountRepository, operationTypeRepo ports.OperationTypeRepository) *CreateTransactionProcessor {
	return &CreateTransactionProcessor{
		transactionRepo:   transactionRepo,
		accountRepo:       accountRepo,
		operationTypeRepo: operationTypeRepo,
	}
}

// Process creates a new transaction with proper amount normalization
func (p *CreateTransactionProcessor) Process(ctx context.Context, req domain.CreateTransactionRequest) (*domain.CreateTransactionResponse, error) {
	// Validate account exists
	account, err := p.accountRepo.FindByID(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("account with id %d does not exist", req.AccountID)
	}

	// Validate operation type exists
	operationType, err := p.operationTypeRepo.FindByID(ctx, req.OperationTypeID)
	if err != nil {
		return nil, fmt.Errorf("operation type not found: %w", err)
	}
	if operationType == nil {
		return nil, domain.ErrInvalidOperationType
	}

	// Create transaction entity
	transaction := &domain.Transaction{
		AccountID:       req.AccountID,
		OperationTypeID: req.OperationTypeID,
		Amount:          req.Amount,
		EventDate:       time.Now().UTC(),
	}

	// Validate transaction
	if err := transaction.Validate(); err != nil {
		return nil, err
	}

	// Normalize amount based on operation type
	if err := transaction.NormalizeAmount(operationType); err != nil {
		return nil, err
	}

	// Save transaction
	createdTransaction, err := p.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Build response
	return &domain.CreateTransactionResponse{
		TransactionID:   createdTransaction.ID,
		AccountID:       createdTransaction.AccountID,
		OperationTypeID: createdTransaction.OperationTypeID,
		Amount:          createdTransaction.Amount,
		EventDate:       createdTransaction.EventDate,
	}, nil
}
