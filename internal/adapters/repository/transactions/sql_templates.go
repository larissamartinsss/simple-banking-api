package transactions

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

	findByAccountIDPaginatedSQL = `SELECT id, account_id, operation_type_id, amount, event_date
		FROM transactions
		WHERE account_id = ?
		ORDER BY event_date DESC
		LIMIT ? OFFSET ?`

	countTransactionsByAccountIDSQL = `
		SELECT COUNT(*)
		FROM transactions
		WHERE account_id = ?
	`

	getAllTransactionsSQL = `
		SELECT id, account_id, operation_type_id, amount, event_date
		FROM transactions
		ORDER BY event_date DESC
	`
)
