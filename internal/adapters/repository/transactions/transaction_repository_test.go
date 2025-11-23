package transactions

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

func setupMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *TransactionRepository) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := NewTransactionRepository(db)
	return db, mock, repo.(*TransactionRepository)
}

func TestCreate(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	now := time.Now()
	input := &domain.Transaction{AccountID: 1, OperationTypeID: 1, Amount: -50.0}

	mock.ExpectQuery("INSERT INTO transactions").
		WithArgs(int64(1), int64(1), -50.0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "operation_type_id", "amount", "event_date"}).
			AddRow(1, 1, 1, -50.0, now))

	result, err := repo.Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, -50.0, result.Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_Error(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	mock.ExpectQuery("INSERT INTO transactions").WillReturnError(sql.ErrConnDone)

	_, err := repo.Create(context.Background(), &domain.Transaction{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create transaction")
}

func TestFindByID(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT (.+) FROM transactions WHERE id").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "operation_type_id", "amount", "event_date"}).
			AddRow(1, 1, 1, -50.0, now))

	result, err := repo.FindByID(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, -50.0, result.Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByID_NotFound(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	mock.ExpectQuery("SELECT (.+) FROM transactions WHERE id").
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.FindByID(context.Background(), 999)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByAccountID(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT (.+) FROM transactions WHERE account_id").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "operation_type_id", "amount", "event_date"}).
			AddRow(1, 1, 1, -50.0, now).
			AddRow(2, 1, 4, 100.0, now))

	results, err := repo.FindByAccountID(context.Background(), 1)

	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, -50.0, results[0].Amount)
	assert.Equal(t, 100.0, results[1].Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByAccountIDPaginated(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	now := time.Now()

	// Mock count query
	mock.ExpectQuery("SELECT COUNT").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	// Mock paginated query
	mock.ExpectQuery("SELECT (.+) FROM transactions WHERE account_id (.+) ORDER BY").
		WithArgs(int64(1), int64(2), int64(0)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "operation_type_id", "amount", "event_date"}).
			AddRow(1, 1, 1, -50.0, now).
			AddRow(2, 1, 4, 100.0, now))

	results, total, err := repo.FindByAccountIDPaginated(context.Background(), 1, 2, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, results, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFindByAccountIDPaginated_CountError(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT").WillReturnError(sql.ErrConnDone)

	_, _, err := repo.FindByAccountIDPaginated(context.Background(), 1, 10, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to count transactions")
}

func TestGetAll(t *testing.T) {
	db, mock, repo := setupMock(t)
	defer db.Close()

	now := time.Now()
	mock.ExpectQuery("SELECT (.+) FROM transactions").
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id", "operation_type_id", "amount", "event_date"}).
			AddRow(1, 1, 1, -50.0, now))

	results, err := repo.GetAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}
