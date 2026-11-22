package main

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/shibukawa/pgmem"
)

var (
	tempServerOnce sync.Once
	tempServer     *pgmem.Server
	tempServerErr  error
)

// openTempDB starts the in-process PostgreSQL (pgmem) and returns an empty
// database. Connections go through pgmem's in-process dialer, not TCP.
//
// The same server is reused within a process; resetTempDB empties it again.
func openTempDB() (*sqlx.DB, error) {
	tempServerOnce.Do(func() {
		tempServer, tempServerErr = pgmem.Start(context.Background(), pgmem.Options{})
	})
	if tempServerErr != nil {
		return nil, tempServerErr
	}

	cfg, err := pgx.ParseConfig(tempServer.DSN())
	if err != nil {
		return nil, err
	}
	cfg.DialFunc = tempServer.Dial
	db := sqlx.NewDb(stdlib.OpenDB(*cfg), "pgx")

	return resetTempDB(db)
}

// resetTempDB drops everything in the database.
func resetTempDB(db *sqlx.DB) (*sqlx.DB, error) {
	if _, err := db.Exec("DROP SCHEMA IF EXISTS public CASCADE"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("CREATE SCHEMA public"); err != nil {
		return nil, err
	}
	return db, nil
}
