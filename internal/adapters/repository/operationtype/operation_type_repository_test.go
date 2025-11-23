package operationtype

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *OperationTypeRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewOperationTypeRepository(db)
	return db, mock, repo.(*OperationTypeRepository)
}

func TestFindByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		mockSetup func(sqlmock.Sqlmock)
		wantFound bool
		wantDesc  string
	}{
		{
			name: "purchase",
			id:   1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM operation_types WHERE id").
					WithArgs(int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "description", "created_at"}).
						AddRow(1, "COMPRA A VISTA", time.Now()))
			},
			wantFound: true,
			wantDesc:  "COMPRA A VISTA",
		},
		{
			name: "credit voucher",
			id:   4,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM operation_types WHERE id").
					WithArgs(int64(4)).
					WillReturnRows(sqlmock.NewRows([]string{"id", "description", "created_at"}).
						AddRow(4, "PAGAMENTO", time.Now()))
			},
			wantFound: true,
			wantDesc:  "PAGAMENTO",
		},
		{
			name: "not found",
			id:   99,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM operation_types WHERE id").
					WithArgs(int64(99)).
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
				assert.Equal(t, tt.wantDesc, result.Description)
			} else {
				assert.Nil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAll(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT (.+) FROM operation_types").
		WillReturnRows(sqlmock.NewRows([]string{"id", "description", "created_at"}).
			AddRow(1, "COMPRA A VISTA", now).
			AddRow(2, "COMPRA PARCELADA", now).
			AddRow(3, "SAQUE", now).
			AddRow(4, "PAGAMENTO", now))

	results, err := repo.GetAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, results, 4)
	assert.Equal(t, "COMPRA A VISTA", results[0].Description)
	assert.Equal(t, int64(1), results[0].ID)
	assert.Equal(t, "PAGAMENTO", results[3].Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}
