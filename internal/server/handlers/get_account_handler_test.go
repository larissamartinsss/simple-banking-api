package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAccountHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		setupMock      func(*mocks.MockGetAccountProcessorInterface)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:      "successfully get account",
			accountID: "1",
			setupMock: func(mockProc *mocks.MockGetAccountProcessorInterface) {
				mockProc.On("Process", mock.Anything, domain.GetAccountRequest{
					AccountID: 1,
				}).Return(&domain.GetAccountResponse{
					Account: &domain.Account{
						ID:             1,
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					},
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result domain.Account
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), result.ID)
				assert.Equal(t, "12345678900", result.DocumentNumber)
			},
		},
		{
			name:      "account not found",
			accountID: "999",
			setupMock: func(mockProc *mocks.MockGetAccountProcessorInterface) {
				mockProc.On("Process", mock.Anything, domain.GetAccountRequest{
					AccountID: 999,
				}).Return(nil, errors.New("account not found")).Once()
			},
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "account not found")
			},
		},
		{
			name:      "invalid account ID format",
			accountID: "invalid",
			setupMock: func(mockProc *mocks.MockGetAccountProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid account ID")
			},
		},
		{
			name:      "zero account ID",
			accountID: "0",
			setupMock: func(mockProc *mocks.MockGetAccountProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "account_id must be greater than 0")
			},
		},
		{
			name:      "negative account ID",
			accountID: "-1",
			setupMock: func(mockProc *mocks.MockGetAccountProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "account_id must be greater than 0")
			},
		},
		{
			name:      "internal server error",
			accountID: "1",
			setupMock: func(mockProc *mocks.MockGetAccountProcessorInterface) {
				mockProc.On("Process", mock.Anything, mock.Anything).
					Return(nil, errors.New("database connection failed")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Failed to retrieve account")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockProc := mocks.NewMockGetAccountProcessorInterface(t)
			if tt.setupMock != nil {
				tt.setupMock(mockProc)
			}

			handler := NewGetAccountHandler(mockProc)

			// Create request with chi context for URL params
			req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+tt.accountID, nil)
			w := httptest.NewRecorder()

			// Add chi URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("accountId", tt.accountID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute
			handler.Handle(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w)
			}

			// Verify mock expectations
			mockProc.AssertExpectations(t)
		})
	}
}
