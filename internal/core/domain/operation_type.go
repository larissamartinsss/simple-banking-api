package domain

import "time"

// OperationType represents the type of transaction operation
type OperationType struct {
	ID          int       `json:"operation_type_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Operation type constants
const (
	OperationTypePurchase                 = 1
	OperationTypePurchaseWithInstallments = 2
	OperationTypeWithdrawal               = 3
	OperationTypeCreditVoucher            = 4
)

// IsDebitOperation checks if the operation type should result in a negative amount
func (ot *OperationType) IsDebitOperation() bool {
	return ot.ID == OperationTypePurchase ||
		ot.ID == OperationTypePurchaseWithInstallments ||
		ot.ID == OperationTypeWithdrawal
}

// IsCreditOperation checks if the operation type should result in a positive amount
func (ot *OperationType) IsCreditOperation() bool {
	return ot.ID == OperationTypeCreditVoucher
}
