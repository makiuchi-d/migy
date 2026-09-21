package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/moriyoshi/pglite-go/pgdriver"
)

// openTempDB starts the in-process PostgreSQL (PGlite) and returns an empty
// database. The wasm assets are embedded at build time (-tags embed).
//
// The cluster only ever has template1, and the driver caches one cluster per
// DSN, so a temporary database is always this one. Starting it takes a couple
// of seconds, so the same one is reused within a process; resetTempDB empties
// it again.
func openTempDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("pglite", "database=template1 ephemeral=true")
	if err != nil {
		return nil, err
	}
	// One embedded backend is one connection.
	db.SetMaxOpenConns(1)

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
