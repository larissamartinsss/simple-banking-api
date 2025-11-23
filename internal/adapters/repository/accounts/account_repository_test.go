package accounts

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *AccountRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewAccountRepository(db)
	return db, mock, repo.(*AccountRepository)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		account     *domain.Account
		mockSetup   func(sqlmock.Sqlmock)
		wantErr     bool
		errContains string
	}{
		{
			name:    "successful creation",
			account: &domain.Account{DocumentNumber: "12345678900"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO accounts").
					WithArgs("12345678900").
					WillReturnRows(sqlmock.NewRows([]string{"id", "document_number", "created_at"}).
						AddRow(1, "12345678900", time.Now()))
			},
			wantErr: false,
		},
		{
			name:    "duplicate document",
			account: &domain.Account{DocumentNumber: "12345678900"},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO accounts").
					WithArgs("12345678900").
					WillReturnError(sql.ErrConnDone)
			},
			wantErr:     true,
			errContains: "failed to create account",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			result, err := repo.Create(context.Background(), tt.account)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.NotZero(t, result.ID)
			assert.Equal(t, tt.account.DocumentNumber, result.DocumentNumber)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFindByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(sqlmock.Sqlmock)
		wantFound bool
		wantErr   bool
	}{
		{
			name: "found",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM accounts WHERE id").
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "document_number", "created_at"}).
						AddRow(1, "12345678900", time.Now()))
			},
			wantFound: true,
		},
		{
			name: "not found",
			id:   999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM accounts WHERE id").
					WithArgs(int64(999)).
					WillReturnError(sql.ErrNoRows)
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			result, err := repo.FindByID(context.Background(), tt.id)

			require.NoError(t, err)
			if tt.wantFound {
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			} else {
				assert.Nil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestFindByDocumentNumber(t *testing.T) {
	tests := []struct {
		name      string
		docNumber string
		mockSetup func(sqlmock.Sqlmock)
		wantFound bool
	}{
		{
			name:      "found",
			docNumber: "12345678900",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM accounts WHERE document_number").
					WithArgs("12345678900").
					WillReturnRows(sqlmock.NewRows([]string{"id", "document_number", "created_at"}).
						AddRow(1, "12345678900", time.Now()))
			},
			wantFound: true,
		},
		{
			name:      "not found",
			docNumber: "99999999999",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM accounts WHERE document_number").
					WithArgs("99999999999").
					WillReturnError(sql.ErrNoRows)
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			result, err := repo.FindByDocumentNumber(context.Background(), tt.docNumber)

			require.NoError(t, err)
			if tt.wantFound {
				assert.NotNil(t, result)
				assert.Equal(t, tt.docNumber, result.DocumentNumber)
			} else {
				assert.Nil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAll(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(sqlmock.Sqlmock)
		wantCount int
	}{
		{
			name: "empty",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM accounts").
					WillReturnRows(sqlmock.NewRows([]string{"id", "document_number", "created_at"}))
			},
			wantCount: 0,
		},
		{
			name: "multiple accounts",
			mockSetup: func(mock sqlmock.Sqlmock) {
				now := time.Now()
				mock.ExpectQuery("SELECT (.+) FROM accounts").
					WillReturnRows(sqlmock.NewRows([]string{"id", "document_number", "created_at"}).
						AddRow(1, "11111111111", now).
						AddRow(2, "22222222222", now).
						AddRow(3, "33333333333", now))
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, repo := setupMock(t)
			defer db.Close()

			tt.mockSetup(mock)

			results, err := repo.GetAll(context.Background())

			require.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
