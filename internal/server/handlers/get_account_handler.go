package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors"
)

type GetAccountHandler struct {
	processor processors.GetAccountProcessorInterface
}

func NewGetAccountHandler(processor processors.GetAccountProcessorInterface) *GetAccountHandler {
	return &GetAccountHandler{
		processor: processor,
	}
}

func (h *GetAccountHandler) Handle(w http.ResponseWriter, r *http.Request) {
	// Extract account ID from URL parameter
	accountIDStr := chi.URLParam(r, "accountId")
	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	// Validate request
	req := domain.GetAccountRequest{
		AccountID: int64(accountID),
	}
	if err := h.validateRequest(req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.processor.Process(r.Context(), req)
	if err != nil {
		if err.Error() == "account not found" {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve account")
		return
	}

	// Return 200 OK with the account
	respondWithJSON(w, http.StatusOK, response.Account)
}

func (h *GetAccountHandler) validateRequest(req domain.GetAccountRequest) error {
	if req.AccountID <= 0 {
		return domain.ErrInvalidAccountID
	}
	return nil
}
