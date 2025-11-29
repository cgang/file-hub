package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var db *bun.DB

func Init(ctx context.Context, dsn string) {
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	if err := pgdb.PingContext(ctx); err != nil {
		log.Panicf("Ping database failed: %s", err)
	}

	db = bun.NewDB(pgdb, pgdialect.New())
}

func Close() error {
	if db == nil {
		return nil
	}

	return db.Close()
}
