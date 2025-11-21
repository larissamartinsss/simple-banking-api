package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

// AccountRepository implements the ports.AccountRepository interface
type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) ports.AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	var result domain.Account

	err := r.db.QueryRowContext(ctx, createAccountSQL, account.DocumentNumber).
		Scan(&result.ID, &result.DocumentNumber, &result.CreatedAt)

	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "UNIQUE constraint failed: accounts.document_number" {
			return nil, errors.New("account with this document number already exists")
		}
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return &result, nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id int64) (*domain.Account, error) {
	var account domain.Account

	err := r.db.QueryRowContext(ctx, findAccountByIDSQL, id).
		Scan(&account.ID, &account.DocumentNumber, &account.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) FindByDocumentNumber(ctx context.Context, documentNumber string) (*domain.Account, error) {
	var account domain.Account

	err := r.db.QueryRowContext(ctx, findAccountByDocumentNumberSQL, documentNumber).
		Scan(&account.ID, &account.DocumentNumber, &account.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) GetAll(ctx context.Context) ([]*domain.Account, error) {
	rows, err := r.db.QueryContext(ctx, getAllAccountsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*domain.Account

	for rows.Next() {
		var account domain.Account
		if err := rows.Scan(&account.ID, &account.DocumentNumber, &account.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating accounts: %w", err)
	}

	return accounts, nil
}
