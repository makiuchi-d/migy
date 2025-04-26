package sqlfile

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Schema struct {
	Tables     []Table
	Procedures []Procedure
	// Triggers
	// Functions
	// Vews
	// Events
}

type Table struct {
	Name   string `db:"Table"`
	Create string `db:"Create Table"`
}

type Procedure struct {
	Name        string `db:"Procedure"`
	Mode        string `db:"sql_mode"`
	Create      string `db:"Create Procedure"`
	Charset     string `db:"character_set_client"`
	Collation   string `db:"collation_connection"`
	DBCollation string `db:"Database Collation"`
}

// GetSchema returns a schema of the current database.
func GetSchema(db *sqlx.DB) (*Schema, error) {
	tbls, err := getTables(db)
	if err != nil {
		return nil, err
	}
	procs, err := getProcedures(db)
	if err != nil {
		return nil, err
	}

	return &Schema{
		Tables:     tbls,
		Procedures: procs,
	}, nil
}

func getTables(db *sqlx.DB) ([]Table, error) {
	var nms []string
	if err := db.Select(&nms, "SHOW TABLES"); err != nil {
		return nil, err
	}
	tbls := make([]Table, 0, len(nms))
	for _, n := range nms {
		var t Table
		err := db.Get(&t, fmt.Sprintf("SHOW CREATE TABLE `%s`", n))
		if err != nil {
			return nil, err
		}
		tbls = append(tbls, t)
	}
	return tbls, nil
}

func getProcedures(db *sqlx.DB) ([]Procedure, error) {
	// Note: go-mysql-server does not currently support "SHOW PROCEDURE STATUS" or "information_schema.routines".
	// This function returns only the required stored procedures.
	// Update this implementation if support is added in the future.
	var nms []string
	nms = append(nms, "_migration_exists")

	procs := make([]Procedure, 0, len(nms))
	for _, n := range nms {
		var p Procedure
		err := db.Get(&p, fmt.Sprintf("SHOW CREATE PROCEDURE `%s`", n))
		if err != nil {
			return nil, err
		}
		procs = append(procs, p)
	}

	return procs, nil
}
