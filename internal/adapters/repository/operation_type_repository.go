package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports"
)

// OperationTypeRepository implements the ports.OperationTypeRepository interface
type OperationTypeRepository struct {
	db *sql.DB
}

// NewOperationTypeRepository creates a new operation type repository
func NewOperationTypeRepository(db *sql.DB) ports.OperationTypeRepository {
	return &OperationTypeRepository{db: db}
}

// FindByID retrieves an operation type by its ID
func (r *OperationTypeRepository) FindByID(ctx context.Context, id int) (*domain.OperationType, error) {
	var opType domain.OperationType

	err := r.db.QueryRowContext(ctx, findOperationTypeByIDSQL, id).
		Scan(&opType.ID, &opType.Description, &opType.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find operation type: %w", err)
	}

	return &opType, nil
}

// GetAll retrieves all operation types
func (r *OperationTypeRepository) GetAll(ctx context.Context) ([]*domain.OperationType, error) {
	rows, err := r.db.QueryContext(ctx, getAllOperationTypesSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation types: %w", err)
	}
	defer rows.Close()

	var operationTypes []*domain.OperationType

	for rows.Next() {
		var opType domain.OperationType
		if err := rows.Scan(&opType.ID, &opType.Description, &opType.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan operation type: %w", err)
		}
		operationTypes = append(operationTypes, &opType)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating operation types: %w", err)
	}

	return operationTypes, nil
}

// Seed initializes the database with the predefined operation types
func (r *OperationTypeRepository) Seed(ctx context.Context) error {
	operationTypes := []struct {
		ID          int
		Description string
	}{
		{domain.OperationTypePurchase, "Normal Purchase"},
		{domain.OperationTypePurchaseWithInstallments, "Purchase with installments"},
		{domain.OperationTypeWithdrawal, "Withdrawal"},
		{domain.OperationTypeCreditVoucher, "Credit Voucher"},
	}

	for _, ot := range operationTypes {
		_, err := r.db.ExecContext(ctx, insertOperationTypeSQL, ot.ID, ot.Description)
		if err != nil {
			return fmt.Errorf("failed to seed operation type %d: %w", ot.ID, err)
		}
	}

	fmt.Println("✅ Seeded operation types")
	return nil
}
