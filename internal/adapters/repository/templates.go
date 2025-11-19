package repository

// SQL queries - Transactions
const (
	createTransactionSQL = `
		INSERT INTO transactions (account_id, operation_type_id, amount, event_date, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, account_id, operation_type_id, amount, event_date
	`

	// Simple query - easy to extend with JOINs later
	// Example: SELECT t.*, m.name as merchant_name FROM transactions t LEFT JOIN merchants m ON t.merchant_id = m.id
	findTransactionByIDSQL = `
		SELECT id, account_id, operation_type_id, amount, event_date
		FROM transactions
		WHERE id = ?
	`

	findTransactionsByAccountIDSQL = `
		SELECT id, account_id, operation_type_id, amount, event_date
		FROM transactions
		WHERE account_id = ?
		ORDER BY event_date DESC
	`

	getAllTransactionsSQL = `
		SELECT id, account_id, operation_type_id, amount, event_date
		FROM transactions
		ORDER BY event_date DESC
	`
)

// SQL queries - OperationTypes
const (
	findOperationTypeByIDSQL = `
		SELECT id, description, created_at
		FROM operation_types
		WHERE id = ?
	`

	getAllOperationTypesSQL = `
		SELECT id, description, created_at
		FROM operation_types
		ORDER BY id
	`

	insertOperationTypeSQL = `
		INSERT OR IGNORE INTO operation_types (id, description, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`
)

// SQL queries - Accounts
const (
	createAccountSQL = `
		INSERT INTO accounts (document_number, created_at)
		VALUES (?, CURRENT_TIMESTAMP)
		RETURNING id, document_number, created_at
	`

	findAccountByIDSQL = `
		SELECT id, document_number, created_at
		FROM accounts
		WHERE id = ?
	`

	findAccountByDocumentNumberSQL = `
		SELECT id, document_number, created_at
		FROM accounts
		WHERE document_number = ?
	`

	getAllAccountsSQL = `
		SELECT id, document_number, created_at
		FROM accounts
		ORDER BY created_at DESC
	`
)
