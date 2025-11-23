package domain

import (
	"errors"
	"regexp"
	"time"
)

// Account errors
var (
	ErrInvalidAccountID = errors.New("account_id must be greater than 0")
)

// Account represents a customer account
type Account struct {
	ID             int64     `json:"account_id"`
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

	// Validate that document_number contains only digits
	matched, err := regexp.MatchString(`^\d+$`, a.DocumentNumber)
	if err != nil {
		return errors.New("failed to validate document_number format")
	}
	if !matched {
		return errors.New("document_number must contain only digits")
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
	AccountID int64 `json:"account_id"`
}

// GetAccountResponse represents the response after getting an account
type GetAccountResponse struct {
	Account *Account `json:"account"`
}
