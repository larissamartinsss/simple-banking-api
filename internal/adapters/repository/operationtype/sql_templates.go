package operationtype

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
