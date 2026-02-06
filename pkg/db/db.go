package db

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

func NewDatabase(ctx context.Context, dn, dsn string) (*sql.DB, error) {
	db, err := sql.Open(dn, dsn)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(5)
	db.SetConnMaxIdleTime(5)

	return db, nil
}
