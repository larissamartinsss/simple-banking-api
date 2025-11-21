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

func TestCreateAccountProcessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.CreateAccountRequest
		setupMocks     func(*mocks.MockAccountRepository)
		wantErr        bool
		wantErrMessage string
		validateResult func(*testing.T, *domain.CreateAccountResponse)
	}{
		{
			name: "successful account creation",
			request: domain.CreateAccountRequest{
				DocumentNumber: "12345678900",
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				// FindByDocumentNumber returns nil (not found)
				mockRepo.EXPECT().
					FindByDocumentNumber(mock.Anything, "12345678900").
					Return(nil, nil).
					Once()

				// Create returns the new account
				mockRepo.EXPECT().
					Create(mock.Anything, mock.MatchedBy(func(acc *domain.Account) bool {
						return acc.DocumentNumber == "12345678900"
					})).
					Return(&domain.Account{
						ID:             int64(1),
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.CreateAccountResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Account)
				assert.Equal(t, int64(1), resp.Account.ID)
				assert.Equal(t, "12345678900", resp.Account.DocumentNumber)
				assert.False(t, resp.Account.CreatedAt.IsZero())
			},
		},
		{
			name: "duplicate document number",
			request: domain.CreateAccountRequest{
				DocumentNumber: "12345678900",
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				// FindByDocumentNumber returns existing account
				mockRepo.EXPECT().
					FindByDocumentNumber(mock.Anything, "12345678900").
					Return(&domain.Account{
						ID:             int64(1),
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					}, nil).
					Once()
				// Create should not be called
			},
			wantErr:        true,
			wantErrMessage: "account with this document number already exists",
		},
		{
			name: "repository find error",
			request: domain.CreateAccountRequest{
				DocumentNumber: "12345678900",
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByDocumentNumber(mock.Anything, "12345678900").
					Return(nil, errors.New("database error")).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "database error",
		},
		{
			name: "repository create error",
			request: domain.CreateAccountRequest{
				DocumentNumber: "12345678900",
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByDocumentNumber(mock.Anything, "12345678900").
					Return(nil, nil).
					Once()

				mockRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(nil, errors.New("failed to insert")).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "failed to insert",
		},
		{
			name: "valid CNPJ document",
			request: domain.CreateAccountRequest{
				DocumentNumber: "12345678901234",
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByDocumentNumber(mock.Anything, "12345678901234").
					Return(nil, nil).
					Once()

				mockRepo.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(&domain.Account{
						ID:             int64(2),
						DocumentNumber: "12345678901234",
						CreatedAt:      time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.CreateAccountResponse) {
				assert.Equal(t, "12345678901234", resp.Account.DocumentNumber)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := mocks.NewMockAccountRepository(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			processor := NewCreateAccountProcessor(mockRepo)
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
			mockRepo.AssertExpectations(t)
		})
	}
}
