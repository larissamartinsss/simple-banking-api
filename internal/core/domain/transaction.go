package domain

import (
	"errors"
	"math"
	"time"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID              int       `json:"transaction_id"`
	AccountID       int       `json:"account_id"`
	OperationTypeID int       `json:"operation_type_id"`
	Amount          float64   `json:"amount"`
	EventDate       time.Time `json:"event_date"`
}

// CreateTransactionRequest represents the input for creating a transaction
type CreateTransactionRequest struct {
	AccountID       int     `json:"account_id"`
	OperationTypeID int     `json:"operation_type_id"`
	Amount          float64 `json:"amount"`
}

// CreateTransactionResponse represents the output after creating a transaction
type CreateTransactionResponse struct {
	TransactionID   int       `json:"transaction_id"`
	AccountID       int       `json:"account_id"`
	OperationTypeID int       `json:"operation_type_id"`
	Amount          float64   `json:"amount"`
	EventDate       time.Time `json:"event_date"`
}

// Validation errors
var (
	ErrInvalidOperationType = errors.New("operation_type_id must be between 1 and 4")
	ErrZeroAmount           = errors.New("amount cannot be zero")
)

// Validate checks if the transaction data is valid
func (t *Transaction) Validate() error {
	if t.AccountID <= 0 {
		return errors.New("account_id must be greater than 0")
	}

	if t.OperationTypeID < 1 || t.OperationTypeID > 4 {
		return errors.New("operation_type_id must be between 1 and 4")
	}

	if t.Amount == 0 {
		return errors.New("amount cannot be zero")
	}

	return nil
}

// NormalizeAmount adjusts the amount sign based on the operation type
// Debit operations (Purchase, Withdrawal) should be negative
// Credit operations (Credit Voucher) should be positive
func (t *Transaction) NormalizeAmount(operationType *OperationType) error {
	if operationType == nil {
		return errors.New("operation type cannot be nil")
	}

	absAmount := math.Abs(t.Amount)

	if operationType.IsDebitOperation() {
		t.Amount = -absAmount
	} else if operationType.IsCreditOperation() {
		t.Amount = absAmount
	}

	return nil
}
