package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// ErrNoRows is returned when a query returns no rows
var ErrNoRows = sql.ErrNoRows

var db *bun.DB

func Init(ctx context.Context, dsn string) {
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	if err := pgdb.PingContext(ctx); err != nil {
		log.Panicf("Ping database failed: %s", err)
	}

	db = bun.NewDB(pgdb, pgdialect.New())
	if err := initDB(ctx); err != nil {
		log.Panicf("Initialize database failed: %s", err)
	}
}

func Close() error {
	if db == nil {
		return nil
	}

	return db.Close()
}

// initDB initializes the database with required tables
func initDB(ctx context.Context) error {
	models := []any{
		(*UserModel)(nil),
		(*FileModel)(nil),
		(*UserQuotaModel)(nil),
	}

	for _, model := range models {
		_, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			return err
		}
	}

	// Create indexes
	for _, query := range []string{
		`CREATE INDEX IF NOT EXISTS idx_files_user_id ON files (user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_files_path ON files (path);`,
	} {
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}

	// Set up the trigger function for automatic quota management
	_, err := db.ExecContext(ctx, `
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
	_, err = db.ExecContext(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'user_quota_trigger') THEN
				CREATE TRIGGER user_quota_trigger
					AFTER INSERT OR UPDATE OR DELETE ON files
					FOR EACH ROW EXECUTE FUNCTION update_user_quota();
			END IF;
		END$$;
	`)
	return err
}
