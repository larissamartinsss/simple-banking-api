package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// GetMigrations returns all migrations in order
// TO ADD NEW COLUMNS: Just add a new migration to this list!
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create initial schema",
			SQL: `
				-- Create operation_types table
				CREATE TABLE IF NOT EXISTS operation_types (
					id INTEGER PRIMARY KEY,
					description TEXT NOT NULL,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				-- Create accounts table
				CREATE TABLE IF NOT EXISTS accounts (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					document_number TEXT NOT NULL UNIQUE,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);

				-- Create transactions table (flexible for future joins!)
				CREATE TABLE IF NOT EXISTS transactions (
					id INTEGER PRIMARY KEY AUTOINCREMENT,
					account_id INTEGER NOT NULL,
					operation_type_id INTEGER NOT NULL,
					amount REAL NOT NULL,
					event_date DATETIME DEFAULT CURRENT_TIMESTAMP,
					created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (account_id) REFERENCES accounts(id),
					FOREIGN KEY (operation_type_id) REFERENCES operation_types(id)
				);

				-- Create indexes for better query performance
				CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
				CREATE INDEX IF NOT EXISTS idx_transactions_operation_type_id ON transactions(operation_type_id);
				CREATE INDEX IF NOT EXISTS idx_accounts_document_number ON accounts(document_number);

				-- Create migration tracking table
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version INTEGER PRIMARY KEY,
					description TEXT NOT NULL,
					applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
				);
			`,
		},
		// EXAMPLE: How to add a new column in the future:
		// {
		// 	Version:     2,
		// 	Description: "Add merchant_id to transactions",
		// 	SQL: `
		// 		ALTER TABLE transactions ADD COLUMN merchant_id INTEGER;
		// 		CREATE INDEX IF NOT EXISTS idx_transactions_merchant_id ON transactions(merchant_id);
		// 	`,
		// },
	}
}

// RunMigrations executes all pending migrations
func RunMigrations(ctx context.Context, db *sql.DB) error {
	migrations := GetMigrations()

	for _, migration := range migrations {
		// Check if migration was already applied
		var count int
		err := db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			migration.Version,
		).Scan(&count)

		// If schema_migrations doesn't exist yet, the first migration will create it
		if err != nil && migration.Version == 1 {
			// Execute first migration which creates the tracking table
			if _, err := db.ExecContext(ctx, migration.SQL); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			}

			// Record the migration
			if _, err := db.ExecContext(ctx,
				"INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)",
				migration.Version, migration.Description, time.Now(),
			); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
			}

			fmt.Printf("✅ Applied migration %d: %s\n", migration.Version, migration.Description)
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			fmt.Printf("⏭️  Skipping migration %d (already applied)\n", migration.Version)
			continue
		}

		// Execute migration
		if _, err := db.ExecContext(ctx, migration.SQL); err != nil {
			return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
		}

		// Record the migration
		if _, err := db.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)",
			migration.Version, migration.Description, time.Now(),
		); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		fmt.Printf("✅ Applied migration %d: %s\n", migration.Version, migration.Description)
	}

	return nil
}
