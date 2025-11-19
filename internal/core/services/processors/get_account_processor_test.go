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

func TestGetAccountProcessor_Process(t *testing.T) {
	tests := []struct {
		name           string
		request        domain.GetAccountRequest
		setupMocks     func(*mocks.MockAccountRepository)
		wantErr        bool
		wantErrMessage string
		validateResult func(*testing.T, *domain.GetAccountResponse)
	}{
		{
			name: "successfully find account",
			request: domain.GetAccountRequest{
				AccountID: 1,
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(&domain.Account{
						ID:             1,
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.GetAccountResponse) {
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Account)
				assert.Equal(t, 1, resp.Account.ID)
				assert.Equal(t, "12345678900", resp.Account.DocumentNumber)
			},
		},
		{
			name: "account not found",
			request: domain.GetAccountRequest{
				AccountID: 999,
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByID(mock.Anything, 999).
					Return(nil, nil).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "account not found",
		},
		{
			name: "repository error",
			request: domain.GetAccountRequest{
				AccountID: 1,
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByID(mock.Anything, 1).
					Return(nil, errors.New("database connection failed")).
					Once()
			},
			wantErr:        true,
			wantErrMessage: "database connection failed",
		},
		{
			name: "find account with large ID",
			request: domain.GetAccountRequest{
				AccountID: 99999,
			},
			setupMocks: func(mockRepo *mocks.MockAccountRepository) {
				mockRepo.EXPECT().
					FindByID(mock.Anything, 99999).
					Return(&domain.Account{
						ID:             99999,
						DocumentNumber: "99988877766",
						CreatedAt:      time.Now(),
					}, nil).
					Once()
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *domain.GetAccountResponse) {
				assert.Equal(t, 99999, resp.Account.ID)
				assert.Equal(t, "99988877766", resp.Account.DocumentNumber)
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

			processor := NewGetAccountProcessor(mockRepo)
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
