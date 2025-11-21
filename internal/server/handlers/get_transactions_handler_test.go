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

func TestGetTransactionsHandler_Handle(t *testing.T) {
	tests := []struct {
		name           string
		accountID      string
		queryParams    string
		setupMock      func(*mocks.MockGetTransactionsProcessorInterface)
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successfully get transactions with default pagination",
			accountID:   "1",
			queryParams: "",
			setupMock: func(mockProc *mocks.MockGetTransactionsProcessorInterface) {
				mockProc.EXPECT().
					Process(mock.Anything, domain.GetTransactionsRequest{
						AccountID: 1,
						Limit:     50,
						Offset:    0,
					}).
					Return(&domain.GetTransactionsResponse{
						Transactions: []*domain.Transaction{
							{
								ID:              1,
								AccountID:       1,
								OperationTypeID: domain.OperationTypePurchase,
								Amount:          -50.0,
								EventDate:       time.Now(),
							},
							{
								ID:              2,
								AccountID:       1,
								OperationTypeID: domain.OperationTypeCreditVoucher,
								Amount:          100.0,
								EventDate:       time.Now(),
							},
						},
						Pagination: domain.PaginationMetadata{
							Total:  2,
							Limit:  50,
							Offset: 0,
							Pages:  1,
						},
					}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result domain.GetTransactionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Len(t, result.Transactions, 2)
				assert.Equal(t, int64(2), result.Pagination.Total)
				assert.Equal(t, int64(1), result.Transactions[0].ID)
			},
		},
		{
			name:        "successfully get transactions with custom pagination",
			accountID:   "1",
			queryParams: "?limit=10&offset=5",
			setupMock: func(mockProc *mocks.MockGetTransactionsProcessorInterface) {
				mockProc.EXPECT().
					Process(mock.Anything, domain.GetTransactionsRequest{
						AccountID: 1,
						Limit:     10,
						Offset:    5,
					}).
					Return(&domain.GetTransactionsResponse{
						Transactions: []*domain.Transaction{},
						Pagination: domain.PaginationMetadata{
							Total:  0,
							Limit:  10,
							Offset: 5,
							Pages:  1,
						},
					}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result domain.GetTransactionsResponse
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Empty(t, result.Transactions)
				assert.Equal(t, int64(10), result.Pagination.Limit)
				assert.Equal(t, int64(5), result.Pagination.Offset)
			},
		},
		{
			name:           "invalid account ID - non-numeric",
			accountID:      "abc",
			queryParams:    "",
			setupMock:      func(mockProc *mocks.MockGetTransactionsProcessorInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid account ID")
			},
		},
		{
			name:           "invalid account ID - zero",
			accountID:      "0",
			queryParams:    "",
			setupMock:      func(mockProc *mocks.MockGetTransactionsProcessorInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid account ID")
			},
		},
		{
			name:           "invalid limit parameter",
			accountID:      "1",
			queryParams:    "?limit=abc",
			setupMock:      func(mockProc *mocks.MockGetTransactionsProcessorInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid limit")
			},
		},
		{
			name:           "invalid offset parameter",
			accountID:      "1",
			queryParams:    "?offset=-5",
			setupMock:      func(mockProc *mocks.MockGetTransactionsProcessorInterface) {},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Invalid offset")
			},
		},
		{
			name:        "account not found",
			accountID:   "999",
			queryParams: "",
			setupMock: func(mockProc *mocks.MockGetTransactionsProcessorInterface) {
				mockProc.EXPECT().
					Process(mock.Anything, domain.GetTransactionsRequest{
						AccountID: 999,
						Limit:     50,
						Offset:    0,
					}).
					Return(nil, errors.New("account with id 999 not found")).
					Once()
			},
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "not found")
			},
		},
		{
			name:        "internal server error",
			accountID:   "1",
			queryParams: "",
			setupMock: func(mockProc *mocks.MockGetTransactionsProcessorInterface) {
				mockProc.EXPECT().
					Process(mock.Anything, domain.GetTransactionsRequest{
						AccountID: 1,
						Limit:     50,
						Offset:    0,
					}).
					Return(nil, errors.New("database error")).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Contains(t, w.Body.String(), "Failed to get transactions")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockProc := mocks.NewMockGetTransactionsProcessorInterface(t)
			tt.setupMock(mockProc)

			// Create handler
			handler := NewGetTransactionsHandler(mockProc)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/accounts/"+tt.accountID+"/transactions"+tt.queryParams, nil)

			// Setup chi context for URL params
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("accountId", tt.accountID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Record response
			w := httptest.NewRecorder()

			// Execute
			handler.Handle(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.validateResp != nil {
				tt.validateResp(t, w)
			}
		})
	}
}
