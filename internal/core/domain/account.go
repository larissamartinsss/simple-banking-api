package domain

import (
	"errors"
	"time"
)

// Account errors
var (
	ErrInvalidAccountID = errors.New("account_id must be greater than 0")
)

// Account represents a customer account
type Account struct {
	ID             int       `json:"account_id"`
	DocumentNumber string    `json:"document_number"`
	CreatedAt      time.Time `json:"created_at"`
}

// Validate checks if the account data is valid
func (a *Account) Validate() error {
	if a.DocumentNumber == "" {
		return errors.New("document_number is required")
	}

	if len(a.DocumentNumber) < 11 || len(a.DocumentNumber) > 14 {
		return errors.New("document_number must have between 11 and 14 characters")
	}

	return nil
}

// CreateAccountRequest represents the request to create an account
type CreateAccountRequest struct {
	DocumentNumber string `json:"document_number"`
}

// CreateAccountResponse represents the response after creating an account
type CreateAccountResponse struct {
	Account *Account `json:"account"`
}

// GetAccountRequest represents the request to get an account
type GetAccountRequest struct {
	AccountID int `json:"account_id"`
}

// GetAccountResponse represents the response after getting an account
type GetAccountResponse struct {
	Account *Account `json:"account"`
}
