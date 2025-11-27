package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// DB provides database access methods
type DB struct {
	*bun.DB
}

// New creates a new database connection
func New(connStr string) (*DB, error) {
	// Create a sql.DB connection using pgdriver
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connStr)))

	// Create a Bun DB instance
	db := bun.NewDB(pgdb, pgdialect.New())

	return &DB{db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.DB.Close()
}

// IsDatabaseEmpty checks if the database is empty (no users exist)
func (d *DB) IsDatabaseEmpty() (bool, error) {
	count, err := d.NewSelect().Model((*User)(nil)).Count(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to count users: %w", err)
	}
	return count == 0, nil
}

// InitDB initializes the database with required tables
func (d *DB) InitDB() error {
	models := []interface{}{
		(*User)(nil),
		(*File)(nil),
		(*UserQuota)(nil),
	}

	for _, model := range models {
		_, err := d.NewCreateTable().Model(model).IfNotExists().Exec(context.Background())
		if err != nil {
			return err
		}
	}

	// Create indexes
	for _, query := range []string{
		`CREATE INDEX IF NOT EXISTS idx_files_user_id ON files (user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_files_path ON files (path);`,
	} {
		_, err := d.ExecContext(context.Background(), query)
		if err != nil {
			return err
		}
	}

	// Set up the trigger function for automatic quota management
	_, err := d.ExecContext(context.Background(), `
		CREATE OR REPLACE FUNCTION update_user_quota()
		RETURNS TRIGGER AS $$
		DECLARE
			v_user_id INTEGER;
			v_size BIGINT;
		BEGIN
			IF (TG_OP = 'INSERT') THEN
				-- When a file is added, increase used quota
				UPDATE user_quota
				SET used_bytes = used_bytes + COALESCE(NEW.size, 0),
					updated_at = CURRENT_TIMESTAMP
				WHERE user_id = NEW.user_id;
				RETURN NEW;
			ELSIF (TG_OP = 'UPDATE') THEN
				-- When a file is updated (size changed), adjust used quota
				UPDATE user_quota
				SET used_bytes = used_bytes - COALESCE(OLD.size, 0) + COALESCE(NEW.size, 0),
					updated_at = CURRENT_TIMESTAMP
				WHERE user_id = NEW.user_id;
				RETURN NEW;
			ELSIF (TG_OP = 'DELETE') THEN
				-- When a file is removed, decrease used quota
				UPDATE user_quota
				SET used_bytes = used_bytes - COALESCE(OLD.size, 0),
					updated_at = CURRENT_TIMESTAMP
				WHERE user_id = OLD.user_id;
				RETURN OLD;
			END IF;
			RETURN NULL;
		END;
		$$ LANGUAGE plpgsql;
	`)
	if err != nil {
		return err
	}

	// Create the trigger if it doesn't exist
	_, err = d.ExecContext(context.Background(), `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'user_quota_trigger') THEN
				CREATE TRIGGER user_quota_trigger
					AFTER INSERT OR UPDATE OR DELETE ON files
					FOR EACH ROW EXECUTE FUNCTION update_user_quota();
			END IF;
		END$$;
	`)
	if err != nil {
		return err
	}

	return nil
}
