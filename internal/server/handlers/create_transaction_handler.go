package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors"
)

type CreateTransactionHandler struct {
	processor processors.CreateTransactionProcessorInterface
}

func NewCreateTransactionHandler(processor processors.CreateTransactionProcessorInterface) *CreateTransactionHandler {
	return &CreateTransactionHandler{
		processor: processor,
	}
}

func (h *CreateTransactionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validateRequest(req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.processor.Process(r.Context(), req)
	if err != nil {
		switch err {
		case domain.ErrInvalidOperationType:
			respondWithError(w, http.StatusBadRequest, err.Error())
		case domain.ErrZeroAmount:
			respondWithError(w, http.StatusBadRequest, err.Error())
		default:
			// Check if it's an account not found error
			errMsg := err.Error()
			if strings.Contains(errMsg, "account not found") ||
				strings.Contains(errMsg, "account with id") ||
				strings.Contains(errMsg, "does not exist") {
				respondWithError(w, http.StatusNotFound, err.Error())
			} else {
				respondWithError(w, http.StatusInternalServerError, "Failed to create transaction")
			}
		}
		return
	}

	// Respond with success
	respondWithJSON(w, http.StatusCreated, response)
}

func (h *CreateTransactionHandler) validateRequest(req domain.CreateTransactionRequest) error {
	if req.AccountID <= 0 {
		return domain.ErrInvalidAccountID
	}

	if req.OperationTypeID < 1 || req.OperationTypeID > 4 {
		return domain.ErrInvalidOperationType
	}

	if req.Amount == 0 {
		return domain.ErrZeroAmount
	}

	return nil
}
