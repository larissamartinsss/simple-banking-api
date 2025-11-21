package processors

import (
	"context"
	"testing"
	"time"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTransactionsProcessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.GetTransactionsRequest
		setupMocks     func(*mocks.MockTransactionRepository, *mocks.MockAccountRepository)
		wantErr        bool
		wantErrMessage string
		validateResult func(*testing.T, *domain.GetTransactionsResponse)
	}{
		{
			name: "successful - get transactions for existing account",
			request: domain.GetTransactionsRequest{
				AccountID: int64(1),
				Limit:     10,
				Offset:    0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, int64(1)).
					Return(&domain.Account{
						ID: int64(1),
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					}, nil).
					Once()

				mockTxRepo.EXPECT().
					FindByAccountIDPaginated(
						mock.Anything, // context
						int64(1),      // accountID
						int64(10),     // limit
						int64(0),      // offset
					).
					Return(
						[]*domain.Transaction{
							{
								ID: int64(1),
								AccountID: int64(1),
								OperationTypeID: domain.OperationTypePurchase,
								Amount:          -50.0,
								EventDate:       time.Now(),
							},
							{
								ID: int64(2),
								AccountID: int64(1),
								OperationTypeID: domain.OperationTypeCreditVoucher,
								Amount:          100.0,
								EventDate:       time.Now(),
							},
						},
						int64(2),
						nil,
					).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.GetTransactionsResponse) {
				assert.NotNil(t, resp, "Response should not be nil")
				assert.Len(t, resp.Transactions, 2, "Should return 2 transactions")

				// Validate pagination
				assert.Equal(t, int64(2), resp.Pagination.Total, "Total should be 2")
				assert.Equal(t, int64(10), resp.Pagination.Limit, "Limit should be 10")
				assert.Equal(t, int64(0), resp.Pagination.Offset, "Offset should be 0")
				assert.Equal(t, int64(1), resp.Pagination.Pages, "Should have 1 page (2 items / 10 limit)")

				// Validate first transaction
				assert.Equal(t, int64(1), resp.Transactions[0].ID)
				assert.Equal(t, -50.0, resp.Transactions[0].Amount)
			},
		},
		{
			name: "error - account not found",
			request: domain.GetTransactionsRequest{
				AccountID: int64(999), // Non-existent account
				Limit:     10,
				Offset:    0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository) {
				// Mock: Account does not exist (returns nil)
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, int64(999)).
					Return(nil, nil). // Returns nil, nil = account not found
					Once()
			},
			wantErr:        true,
			wantErrMessage: "account with id 999 not found",
			validateResult: nil, // Does not validate result when there is an error
		},
		{
			name: "successful - account with no transactions",
			request: domain.GetTransactionsRequest{
				AccountID: int64(1),
				Limit:     50,
				Offset:    0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, int64(1)).
					Return(&domain.Account{
						ID: int64(1),
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					}, nil).
					Once()

					// Fetch transactions returns empty list
				mockTxRepo.EXPECT().
					FindByAccountIDPaginated(mock.Anything, int64(1), int64(50), int64(0)).
					Return(
						[]*domain.Transaction{}, // Empty list
						int64(0),                // Total = 0
						nil,
					).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.GetTransactionsResponse) {
				assert.NotNil(t, resp)
				assert.Empty(t, resp.Transactions, "List should be empty")
				assert.Equal(t, int64(0), resp.Pagination.Total)
				assert.Equal(t, int64(1), resp.Pagination.Pages, "There is always at least 1 page")
			},
		},
	}

	//LOOP: Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1️⃣ ARRANGE: Prepare the mocks
			mockTxRepo := mocks.NewMockTransactionRepository(t)
			mockAccRepo := mocks.NewMockAccountRepository(t)

			tt.setupMocks(mockTxRepo, mockAccRepo)

			processor := NewGetTransactionsProcessor(mockTxRepo, mockAccRepo)

			// 2️ACT: Execute the action
			result, err := processor.Process(context.Background(), tt.request)

			if tt.wantErr {
				// We expect an error
				assert.Error(t, err, "Should return an error")
				if tt.wantErrMessage != "" {
					assert.Contains(t, err.Error(), tt.wantErrMessage)
				}
				assert.Nil(t, result, "Result should be nil when there is an error")
			} else {
				// We do not expect an error
				assert.NoError(t, err, "Should not return an error")
				assert.NotNil(t, result, "Result should not be nil")
				// Custom validations
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}
