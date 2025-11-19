package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

// TransactionRepository implements the ports.TransactionRepository interface
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) ports.TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) (*domain.Transaction, error) {
	var result domain.Transaction

	err := r.db.QueryRowContext(
		ctx,
		createTransactionSQL,
		transaction.AccountID,
		transaction.OperationTypeID,
		transaction.Amount,
	).Scan(
		&result.ID,
		&result.AccountID,
		&result.OperationTypeID,
		&result.Amount,
		&result.EventDate,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return &result, nil
}

// FindByID retrieves a transaction by its ID
func (r *TransactionRepository) FindByID(ctx context.Context, id int) (*domain.Transaction, error) {
	var transaction domain.Transaction

	err := r.db.QueryRowContext(ctx, findTransactionByIDSQL, id).
		Scan(
			&transaction.ID,
			&transaction.AccountID,
			&transaction.OperationTypeID,
			&transaction.Amount,
			&transaction.EventDate,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find transaction: %w", err)
	}

	return &transaction, nil
}

// FindByAccountID retrieves all transactions for a specific account
func (r *TransactionRepository) FindByAccountID(ctx context.Context, accountID int) ([]*domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, findTransactionsByAccountIDSQL, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

// GetAll retrieves all transactions
func (r *TransactionRepository) GetAll(ctx context.Context) ([]*domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx, getAllTransactionsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

// scanTransactions is a helper to scan multiple transactions
// When adding new columns, just update this method!
func (r *TransactionRepository) scanTransactions(rows *sql.Rows) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	for rows.Next() {
		var transaction domain.Transaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.AccountID,
			&transaction.OperationTypeID,
			&transaction.Amount,
			&transaction.EventDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, &transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}
