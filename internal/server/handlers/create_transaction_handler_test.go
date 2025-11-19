package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTransactionHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockCreateTransactionProcessorInterface)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful purchase transaction",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 1,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				mockProc.On("Process", mock.Anything, domain.CreateTransactionRequest{
					AccountID:       1,
					OperationTypeID: 1,
					Amount:          50.0,
				}).Return(&domain.CreateTransactionResponse{
					TransactionID:   1,
					AccountID:       1,
					OperationTypeID: 1,
					Amount:          -50.0, // Normalized to negative
					EventDate:       time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result domain.CreateTransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, 1, result.TransactionID)
				assert.Equal(t, 1, result.AccountID)
				assert.Equal(t, -50.0, result.Amount)
			},
		},
		{
			name: "successful credit voucher",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 4,
				"amount":            100.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				mockProc.On("Process", mock.Anything, domain.CreateTransactionRequest{
					AccountID:       1,
					OperationTypeID: 4,
					Amount:          100.0,
				}).Return(&domain.CreateTransactionResponse{
					TransactionID:   2,
					AccountID:       1,
					OperationTypeID: 4,
					Amount:          100.0, // Stays positive
					EventDate:       time.Now(),
				}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result domain.CreateTransactionResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, 100.0, result.Amount)
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid json",
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				// No mock expectations as it should fail before calling processor
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid request body")
			},
		},
		{
			name: "zero account ID",
			requestBody: map[string]interface{}{
				"account_id":        0,
				"operation_type_id": 1,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "account_id must be greater than 0")
			},
		},
		{
			name: "negative account ID",
			requestBody: map[string]interface{}{
				"account_id":        -1,
				"operation_type_id": 1,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "account_id must be greater than 0")
			},
		},
		{
			name: "invalid operation type (too high)",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 99,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "operation_type_id must be between 1 and 4")
			},
		},
		{
			name: "invalid operation type (zero)",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 0,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "operation_type_id must be between 1 and 4")
			},
		},
		{
			name: "zero amount",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 1,
				"amount":            0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				// No mock expectations as validation should fail
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "amount cannot be zero")
			},
		},
		{
			name: "account not found",
			requestBody: map[string]interface{}{
				"account_id":        999,
				"operation_type_id": 1,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				mockProc.On("Process", mock.Anything, mock.Anything).
					Return(nil, errors.New("account with id 999 does not exist")).
					Once()
			},
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				// The handler checks if error starts with "account with id"
				assert.Contains(t, w.Body.String(), "account with id 999")
			},
		},
		{
			name: "invalid operation type from processor",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 3,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				mockProc.On("Process", mock.Anything, mock.Anything).
					Return(nil, domain.ErrInvalidOperationType).
					Once()
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "operation_type_id must be between 1 and 4")
			},
		},
		{
			name: "internal server error",
			requestBody: map[string]interface{}{
				"account_id":        1,
				"operation_type_id": 1,
				"amount":            50.0,
			},
			setupMock: func(mockProc *mocks.MockCreateTransactionProcessorInterface) {
				mockProc.On("Process", mock.Anything, mock.Anything).
					Return(nil, errors.New("database connection failed")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Failed to create transaction")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockProc := mocks.NewMockCreateTransactionProcessorInterface(t)
			if tt.setupMock != nil {
				tt.setupMock(mockProc)
			}

			handler := NewCreateTransactionHandler(mockProc)

			// Create request
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

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
