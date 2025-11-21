package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors"
)

type GetTransactionsHandler struct {
	processor processors.GetTransactionsProcessorInterface
}

func NewGetTransactionsHandler(processor processors.GetTransactionsProcessorInterface) *GetTransactionsHandler {
	return &GetTransactionsHandler{
		processor: processor,
	}
}

func (h *GetTransactionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	accountIDStr := chi.URLParam(r, "accountId")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil || accountID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	// Get pagination parameters from query string
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Default values
	limit := 50
	offset := 0

	// Parse limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid limit")
			return
		}
		limit = parsedLimit
	}

	// Parse offset
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			respondWithError(w, http.StatusBadRequest, "Invalid offset")
			return
		}
		offset = parsedOffset
	}

	req := domain.GetTransactionsRequest{
		AccountID: accountID,
		Limit:     int64(limit),
		Offset:    int64(offset),
	}

	response, err := h.processor.Process(r.Context(), req)
	if err != nil {
		if contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get transactions")
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
