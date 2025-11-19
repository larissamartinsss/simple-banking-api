package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	_ "modernc.org/sqlite"
)

// setupOperationTypeTestDB creates an in-memory SQLite database for testing operation types
func setupOperationTypeTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS operation_types (
			id INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestOperationTypeRepository_Seed(t *testing.T) {
	tests := []struct {
		name          string
		setupData     func(*sql.DB)
		wantErr       bool
		expectedCount int
	}{
		{
			name:          "seed empty database",
			setupData:     func(db *sql.DB) {},
			wantErr:       false,
			expectedCount: 4,
		},
		{
			name: "seed already seeded database (idempotent)",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT OR REPLACE INTO operation_types (id, description) VALUES (1, 'Normal Purchase')")
				db.Exec("INSERT OR REPLACE INTO operation_types (id, description) VALUES (2, 'Purchase with installments')")
			},
			wantErr:       false,
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupOperationTypeTestDB(t)
			defer db.Close()

			if tt.setupData != nil {
				tt.setupData(db)
			}

			repo := NewOperationTypeRepository(db)
			ctx := context.Background()

			err := repo.Seed(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Seed() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Seed() unexpected error = %v", err)
				return
			}

			// Verify all 4 operation types were seeded
			var count int
			db.QueryRow("SELECT COUNT(*) FROM operation_types").Scan(&count)
			if count != tt.expectedCount {
				t.Errorf("Seed() operation type count = %v, want %v", count, tt.expectedCount)
			}

			// Verify specific operation types exist
			expectedTypes := map[int]string{
				domain.OperationTypePurchase:                 "Normal Purchase",
				domain.OperationTypePurchaseWithInstallments: "Purchase with installments",
				domain.OperationTypeWithdrawal:               "Withdrawal",
				domain.OperationTypeCreditVoucher:            "Credit Voucher",
			}

			for id, expectedDesc := range expectedTypes {
				var desc string
				err := db.QueryRow("SELECT description FROM operation_types WHERE id = ?", id).Scan(&desc)
				if err != nil {
					t.Errorf("Seed() operation type %d not found", id)
					continue
				}
				if desc != expectedDesc {
					t.Errorf("Seed() operation type %d description = %v, want %v", id, desc, expectedDesc)
				}
			}
		})
	}
}

func TestOperationTypeRepository_FindByID(t *testing.T) {
	tests := []struct {
		name      string
		opTypeID  int
		setupData func(*sql.DB)
		wantFound bool
		wantDesc  string
		wantErr   bool
	}{
		{
			name:     "find purchase operation",
			opTypeID: 1,
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (1, 'Normal Purchase')")
			},
			wantFound: true,
			wantDesc:  "Normal Purchase",
			wantErr:   false,
		},
		{
			name:     "find credit voucher operation",
			opTypeID: 4,
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (1, 'Normal Purchase')")
				db.Exec("INSERT INTO operation_types (id, description) VALUES (4, 'Credit Voucher')")
			},
			wantFound: true,
			wantDesc:  "Credit Voucher",
			wantErr:   false,
		},
		{
			name:     "operation type not found",
			opTypeID: 999,
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (1, 'Normal Purchase')")
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:     "find withdrawal operation",
			opTypeID: 3,
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (3, 'Withdrawal')")
			},
			wantFound: true,
			wantDesc:  "Withdrawal",
			wantErr:   false,
		},
		{
			name:     "find purchase with installments",
			opTypeID: 2,
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (2, 'Purchase with installments')")
			},
			wantFound: true,
			wantDesc:  "Purchase with installments",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupOperationTypeTestDB(t)
			defer db.Close()

			if tt.setupData != nil {
				tt.setupData(db)
			}

			repo := NewOperationTypeRepository(db)
			ctx := context.Background()

			result, err := repo.FindByID(ctx, tt.opTypeID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindByID() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FindByID() unexpected error = %v", err)
				return
			}

			if tt.wantFound {
				if result == nil {
					t.Error("FindByID() expected operation type but got nil")
					return
				}
				if result.ID != tt.opTypeID {
					t.Errorf("FindByID() ID = %v, want %v", result.ID, tt.opTypeID)
				}
				if result.Description != tt.wantDesc {
					t.Errorf("FindByID() description = %v, want %v", result.Description, tt.wantDesc)
				}
				if result.CreatedAt.IsZero() {
					t.Error("FindByID() returned zero CreatedAt timestamp")
				}
			} else {
				if result != nil {
					t.Errorf("FindByID() expected nil but got operation type: %+v", result)
				}
			}
		})
	}
}

func TestOperationTypeRepository_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		setupData func(*sql.DB)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "empty database",
			setupData: func(db *sql.DB) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "single operation type",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (1, 'Normal Purchase')")
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "all four operation types",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO operation_types (id, description) VALUES (1, 'Normal Purchase')")
				db.Exec("INSERT INTO operation_types (id, description) VALUES (2, 'Purchase with installments')")
				db.Exec("INSERT INTO operation_types (id, description) VALUES (3, 'Withdrawal')")
				db.Exec("INSERT INTO operation_types (id, description) VALUES (4, 'Credit Voucher')")
			},
			wantCount: 4,
			wantErr:   false,
		},
		{
			name: "seeded database",
			setupData: func(db *sql.DB) {
				repo := NewOperationTypeRepository(db)
				repo.Seed(context.Background())
			},
			wantCount: 4,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupOperationTypeTestDB(t)
			defer db.Close()

			if tt.setupData != nil {
				tt.setupData(db)
			}

			repo := NewOperationTypeRepository(db)
			ctx := context.Background()

			results, err := repo.GetAll(ctx)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetAll() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetAll() unexpected error = %v", err)
				return
			}

			if len(results) != tt.wantCount {
				t.Errorf("GetAll() count = %v, want %v", len(results), tt.wantCount)
			}

			// Validate each operation type has required fields
			for i, opType := range results {
				if opType.ID == 0 {
					t.Errorf("GetAll() operation_type[%d] has zero ID", i)
				}
				if opType.Description == "" {
					t.Errorf("GetAll() operation_type[%d] has empty Description", i)
				}
				if opType.CreatedAt.IsZero() {
					t.Errorf("GetAll() operation_type[%d] has zero CreatedAt", i)
				}
			}
		})
	}
}

func TestOperationTypeRepository_OperationTypeConstants(t *testing.T) {
	tests := []struct {
		name             string
		operationID      int
		expectedIsDebit  bool
		expectedIsCredit bool
	}{
		{
			name:             "purchase is debit",
			operationID:      domain.OperationTypePurchase,
			expectedIsDebit:  true,
			expectedIsCredit: false,
		},
		{
			name:             "purchase with installments is debit",
			operationID:      domain.OperationTypePurchaseWithInstallments,
			expectedIsDebit:  true,
			expectedIsCredit: false,
		},
		{
			name:             "withdrawal is debit",
			operationID:      domain.OperationTypeWithdrawal,
			expectedIsDebit:  true,
			expectedIsCredit: false,
		},
		{
			name:             "credit voucher is credit",
			operationID:      domain.OperationTypeCreditVoucher,
			expectedIsDebit:  false,
			expectedIsCredit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupOperationTypeTestDB(t)
			defer db.Close()

			// Seed database
			repo := NewOperationTypeRepository(db)
			ctx := context.Background()
			repo.Seed(ctx)

			// Find the operation type
			opType, err := repo.FindByID(ctx, tt.operationID)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}
			if opType == nil {
				t.Fatalf("FindByID() returned nil for operation type %d", tt.operationID)
			}

			// Test IsDebitOperation
			if opType.IsDebitOperation() != tt.expectedIsDebit {
				t.Errorf("IsDebitOperation() = %v, want %v", opType.IsDebitOperation(), tt.expectedIsDebit)
			}

			// Test IsCreditOperation
			if opType.IsCreditOperation() != tt.expectedIsCredit {
				t.Errorf("IsCreditOperation() = %v, want %v", opType.IsCreditOperation(), tt.expectedIsCredit)
			}
		})
	}
}
