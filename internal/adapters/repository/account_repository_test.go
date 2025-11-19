package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	_ "modernc.org/sqlite"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create schema
	schema := `
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			document_number TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestAccountRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		account     *domain.Account
		wantErr     bool
		errContains string
		setupData   func(*sql.DB) // pre-populate data
	}{
		{
			name: "successful account creation",
			account: &domain.Account{
				DocumentNumber: "12345678900",
			},
			wantErr: false,
		},
		{
			name: "duplicate document number",
			account: &domain.Account{
				DocumentNumber: "99988877766",
			},
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "99988877766")
			},
			wantErr:     true,
			errContains: "UNIQUE constraint failed",
		},
		{
			name: "valid CPF format",
			account: &domain.Account{
				DocumentNumber: "11122233344",
			},
			wantErr: false,
		},
		{
			name: "valid CNPJ format",
			account: &domain.Account{
				DocumentNumber: "12345678901234",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			if tt.setupData != nil {
				tt.setupData(db)
			}

			repo := NewAccountRepository(db)
			ctx := context.Background()

			result, err := repo.Create(ctx, tt.account)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error but got none")
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Create() error = %v, should contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Error("Create() returned nil result")
				return
			}

			if result.ID == 0 {
				t.Error("Create() returned account with zero ID")
			}

			if result.DocumentNumber != tt.account.DocumentNumber {
				t.Errorf("Create() document_number = %v, want %v", result.DocumentNumber, tt.account.DocumentNumber)
			}

			if result.CreatedAt.IsZero() {
				t.Error("Create() returned zero CreatedAt timestamp")
			}
		})
	}
}

func TestAccountRepository_FindByID(t *testing.T) {
	tests := []struct {
		name      string
		accountID int
		setupData func(*sql.DB) int // returns the ID to search for
		wantFound bool
		wantErr   bool
	}{
		{
			name:      "find existing account",
			accountID: 0, // will be set by setupData
			setupData: func(db *sql.DB) int {
				result, _ := db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "12345678900")
				id, _ := result.LastInsertId()
				return int(id)
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:      "account not found",
			accountID: 999,
			setupData: func(db *sql.DB) int {
				return 999
			},
			wantFound: false,
			wantErr:   false,
		},
		{
			name:      "find account by ID 1",
			accountID: 0,
			setupData: func(db *sql.DB) int {
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "11111111111")
				result, _ := db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "22222222222")
				id, _ := result.LastInsertId()
				return int(id)
			},
			wantFound: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			searchID := tt.accountID
			if tt.setupData != nil {
				searchID = tt.setupData(db)
			}

			repo := NewAccountRepository(db)
			ctx := context.Background()

			result, err := repo.FindByID(ctx, searchID)

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
					t.Error("FindByID() expected account but got nil")
					return
				}
				if result.ID != searchID {
					t.Errorf("FindByID() ID = %v, want %v", result.ID, searchID)
				}
			} else {
				if result != nil {
					t.Errorf("FindByID() expected nil but got account: %+v", result)
				}
			}
		})
	}
}

func TestAccountRepository_FindByDocumentNumber(t *testing.T) {
	tests := []struct {
		name           string
		documentNumber string
		setupData      func(*sql.DB)
		wantFound      bool
		wantErr        bool
	}{
		{
			name:           "find existing account by document",
			documentNumber: "12345678900",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "12345678900")
			},
			wantFound: true,
			wantErr:   false,
		},
		{
			name:           "document not found",
			documentNumber: "00000000000",
			setupData:      func(db *sql.DB) {},
			wantFound:      false,
			wantErr:        false,
		},
		{
			name:           "find second account",
			documentNumber: "99988877766",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "11111111111")
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "99988877766")
			},
			wantFound: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			if tt.setupData != nil {
				tt.setupData(db)
			}

			repo := NewAccountRepository(db)
			ctx := context.Background()

			result, err := repo.FindByDocumentNumber(ctx, tt.documentNumber)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FindByDocumentNumber() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FindByDocumentNumber() unexpected error = %v", err)
				return
			}

			if tt.wantFound {
				if result == nil {
					t.Error("FindByDocumentNumber() expected account but got nil")
					return
				}
				if result.DocumentNumber != tt.documentNumber {
					t.Errorf("FindByDocumentNumber() document = %v, want %v", result.DocumentNumber, tt.documentNumber)
				}
			} else {
				if result != nil {
					t.Errorf("FindByDocumentNumber() expected nil but got account: %+v", result)
				}
			}
		})
	}
}

func TestAccountRepository_GetAll(t *testing.T) {
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
			name: "single account",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "12345678900")
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "multiple accounts",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "11111111111")
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "22222222222")
				db.Exec("INSERT INTO accounts (document_number) VALUES (?)", "33333333333")
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "accounts with different timestamps",
			setupData: func(db *sql.DB) {
				db.Exec("INSERT INTO accounts (document_number, created_at) VALUES (?, ?)", "11111111111", time.Now().Add(-1*time.Hour))
				db.Exec("INSERT INTO accounts (document_number, created_at) VALUES (?, ?)", "22222222222", time.Now())
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			if tt.setupData != nil {
				tt.setupData(db)
			}

			repo := NewAccountRepository(db)
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

			// Validate each account has required fields
			for i, account := range results {
				if account.ID == 0 {
					t.Errorf("GetAll() account[%d] has zero ID", i)
				}
				if account.DocumentNumber == "" {
					t.Errorf("GetAll() account[%d] has empty DocumentNumber", i)
				}
				if account.CreatedAt.IsZero() {
					t.Errorf("GetAll() account[%d] has zero CreatedAt", i)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
