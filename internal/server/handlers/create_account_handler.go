package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/larissamartinsss/simple-banking-api/internal/core/domain"
	"github.com/larissamartinsss/simple-banking-api/internal/core/services/processors"
)

type CreateAccountHandler struct {
	processor processors.CreateAccountProcessorInterface
}

func NewCreateAccountHandler(processor processors.CreateAccountProcessorInterface) *CreateAccountHandler {
	return &CreateAccountHandler{
		processor: processor,
	}
}

func (h *CreateAccountHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAccountRequest
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
		if err.Error() == "account with this document number already exists" {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}

	respondWithJSON(w, http.StatusCreated, response.Account)
}

func (h *CreateAccountHandler) validateRequest(req domain.CreateAccountRequest) error {
	account := &domain.Account{
		DocumentNumber: req.DocumentNumber,
	}
	return account.Validate()
}
