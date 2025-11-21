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

func TestCreateAccountHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mocks.MockCreateAccountProcessorInterface)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful account creation",
			requestBody: map[string]string{
				"document_number": "12345678900",
			},
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
				mockProc.On("Process", mock.Anything, domain.CreateAccountRequest{
					DocumentNumber: "12345678900",
				}).Return(&domain.CreateAccountResponse{
					Account: &domain.Account{
						ID:             1,
						DocumentNumber: "12345678900",
						CreatedAt:      time.Now(),
					},
				}, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result domain.Account
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), result.ID)
				assert.Equal(t, "12345678900", result.DocumentNumber)
			},
		},
		{
			name:        "invalid JSON body",
			requestBody: "invalid json",
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid request body")
			},
		},
		{
			name: "empty document number",
			requestBody: map[string]string{
				"document_number": "",
			},
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "document_number is required")
			},
		},
		{
			name: "document number too short",
			requestBody: map[string]string{
				"document_number": "123",
			},
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "document_number must have between 11 and 14 characters")
			},
		},
		{
			name: "document number too long",
			requestBody: map[string]string{
				"document_number": "123456789012345",
			},
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "document_number must have between 11 and 14 characters")
			},
		},
		{
			name: "duplicate document number",
			requestBody: map[string]string{
				"document_number": "12345678900",
			},
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
				mockProc.On("Process", mock.Anything, mock.Anything).
					Return(nil, errors.New("account with this document number already exists")).
					Once()
			},
			expectedStatus: http.StatusConflict,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "account with this document number already exists")
			},
		},
		{
			name: "internal server error",
			requestBody: map[string]string{
				"document_number": "12345678900",
			},
			setupMock: func(mockProc *mocks.MockCreateAccountProcessorInterface) {
				mockProc.On("Process", mock.Anything, mock.Anything).
					Return(nil, errors.New("database connection failed")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Failed to create account")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProc := mocks.NewMockCreateAccountProcessorInterface(t)
			if tt.setupMock != nil {
				tt.setupMock(mockProc)
			}

			handler := NewCreateAccountHandler(mockProc)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Handle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w)
			}
			mockProc.AssertExpectations(t)
		})
	}
}
