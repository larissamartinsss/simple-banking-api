package accounts

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
