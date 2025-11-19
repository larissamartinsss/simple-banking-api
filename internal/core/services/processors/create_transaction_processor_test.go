package processors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/ports/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransactionProcessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.CreateTransactionRequest
		setupMocks     func(*mocks.MockTransactionRepository, *mocks.MockAccountRepository, *mocks.MockOperationTypeRepository)
		wantErr        bool
		wantErrMessage string
		validateResult func(*testing.T, *domain.CreateTransactionResponse)
	}{
		{
			name: "successful purchase transaction (negative amount)",
			request: domain.CreateTransactionRequest{
				AccountID:       1,
				OperationTypeID: domain.OperationTypePurchase,
				Amount:          50.0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository, mockOpRepo *mocks.MockOperationTypeRepository) {
				// Account exists
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(&domain.Account{
						ID:             1,
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					}, nil).
					Once()

				// Operation type exists (Purchase)
				mockOpRepo.EXPECT().
					FindByID(mock.Anything, domain.OperationTypePurchase).
					Return(&domain.OperationType{
						ID:          domain.OperationTypePurchase,
						Description: "Normal Purchase",
						CreatedAt:   time.Now(),
					}, nil).
					Once()

				// Transaction created with negative amount
				mockTxRepo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(tx *domain.Transaction) bool {
						return tx.AccountID == 1 &&
							tx.OperationTypeID == domain.OperationTypePurchase &&
							tx.Amount == -50.0 // Should be negative
					})).
					Return(&domain.Transaction{
						ID:              1,
						AccountID:       1,
						OperationTypeID: domain.OperationTypePurchase,
						Amount:          -50.0,
						EventDate:       time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.CreateTransactionResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, 1, resp.TransactionID)
				assert.Equal(t, 1, resp.AccountID)
				assert.Equal(t, domain.OperationTypePurchase, resp.OperationTypeID)
				assert.Equal(t, -50.0, resp.Amount) // Normalized to negative
			},
		},
		{
			name: "successful credit voucher (positive amount)",
			request: domain.CreateTransactionRequest{
				AccountID:       1,
				OperationTypeID: domain.OperationTypeCreditVoucher,
				Amount:          -100.0, // Sending negative but should be corrected
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository, mockOpRepo *mocks.MockOperationTypeRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(&domain.Account{ID: 1, DocumentNumber: "12345678900"}, nil).
					Once()

				mockOpRepo.EXPECT().
					FindByID(mock.Anything, domain.OperationTypeCreditVoucher).
					Return(&domain.OperationType{
						ID:          domain.OperationTypeCreditVoucher,
						Description: "Credit Voucher",
					}, nil).
					Once()

				mockTxRepo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(tx *domain.Transaction) bool {
						return tx.Amount == 100.0 // Should be positive
					})).
					Return(&domain.Transaction{
						ID:              2,
						AccountID:       1,
						OperationTypeID: domain.OperationTypeCreditVoucher,
						Amount:          100.0,
						EventDate:       time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.CreateTransactionResponse) {
				assert.Equal(t, 100.0, resp.Amount) // Normalized to positive
			},
		},
		{
			name: "account not found",
			request: domain.CreateTransactionRequest{
				AccountID:       999,
				OperationTypeID: domain.OperationTypePurchase,
				Amount:          50.0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository, mockOpRepo *mocks.MockOperationTypeRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, 999).
					Return(nil, nil).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "account with id 999 does not exist",
		},
		{
			name: "invalid operation type",
			request: domain.CreateTransactionRequest{
				AccountID:       1,
				OperationTypeID: 99,
				Amount:          50.0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository, mockOpRepo *mocks.MockOperationTypeRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(&domain.Account{ID: 1, DocumentNumber: "12345678900"}, nil).
					Once()

				mockOpRepo.EXPECT().
					FindByID(mock.Anything, 99).
					Return(nil, nil).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "operation_type_id must be between 1 and 4",
		},
		{
			name: "withdrawal transaction (negative)",
			request: domain.CreateTransactionRequest{
				AccountID:       1,
				OperationTypeID: domain.OperationTypeWithdrawal,
				Amount:          30.0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository, mockOpRepo *mocks.MockOperationTypeRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(&domain.Account{ID: 1}, nil).
					Once()

				mockOpRepo.EXPECT().
					FindByID(mock.Anything, domain.OperationTypeWithdrawal).
					Return(&domain.OperationType{
						ID:          domain.OperationTypeWithdrawal,
						Description: "Withdrawal",
					}, nil).
					Once()

				mockTxRepo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(tx *domain.Transaction) bool {
						return tx.Amount == -30.0 // Should be negative
					})).
					Return(&domain.Transaction{
						ID:              3,
						AccountID:       1,
						OperationTypeID: domain.OperationTypeWithdrawal,
						Amount:          -30.0,
						EventDate:       time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.CreateTransactionResponse) {
				assert.Equal(t, -30.0, resp.Amount)
			},
		},
		{
			name: "repository create error",
			request: domain.CreateTransactionRequest{
				AccountID:       1,
				OperationTypeID: domain.OperationTypePurchase,
				Amount:          50.0,
			},
			setupMocks: func(mockTxRepo *mocks.MockTransactionRepository, mockAccRepo *mocks.MockAccountRepository, mockOpRepo *mocks.MockOperationTypeRepository) {
				mockAccRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(&domain.Account{ID: 1}, nil).
					Once()

				mockOpRepo.EXPECT().
					FindByID(mock.Anything, domain.OperationTypePurchase).
					Return(&domain.OperationType{ID: domain.OperationTypePurchase, Description: "Normal Purchase"}, nil).
					Once()

				mockTxRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil, errors.New("database insert failed")).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "failed to create transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockTxRepo := mocks.NewMockTransactionRepository(t)
			mockAccRepo := mocks.NewMockAccountRepository(t)
			mockOpRepo := mocks.NewMockOperationTypeRepository(t)

			if tt.setupMocks != nil {
				tt.setupMocks(mockTxRepo, mockAccRepo, mockOpRepo)
			}

			processor := NewCreateTransactionProcessor(mockTxRepo, mockAccRepo, mockOpRepo)
			ctx := context.Background()

			// Execute
			result, err := processor.Process(ctx, tt.request)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMessage != "" {
					assert.Contains(t, err.Error(), tt.wantErrMessage)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}

			// Verify all expectations were met
			mockTxRepo.AssertExpectations(t)
			mockAccRepo.AssertExpectations(t)
			mockOpRepo.AssertExpectations(t)
		})
	}
}
