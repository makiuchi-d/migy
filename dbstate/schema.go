package dbstate

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

var ErrNoMigrationTable = errors.New("no '_migrations' table")

type Table struct {
	Name   string `db:"Table"`
	Create string `db:"Create Table"`
	Refs   []string
}

type Procedure struct {
	Name        string `db:"Procedure"`
	Mode        string `db:"sql_mode"`
	Create      string `db:"Create Procedure"`
	Charset     string `db:"character_set_client"`
	Collation   string `db:"collation_connection"`
	DBCollation string `db:"Database Collation"`
}

var reRef = regexp.MustCompile("REFERENCES `([^`]*)`")

func HasMigrationTable(db *sqlx.DB) error {
	const q = "SHOW TABLES LIKE '_migrations'"
	var s string
	err := db.Get(&s, q)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoMigrationTable
	}
	return err
}

// GetTables returns table informations
func GetTables(db *sqlx.DB) ([]*Table, error) {
	var names []string
	if err := db.Select(&names, "SHOW TABLES"); err != nil {
		return nil, err
	}
	tbls := make([]*Table, 0, len(names))
	done := make(map[string]struct{})
	pendings := make(map[string][]*Table)
	var appendTable func(t *Table) bool
	appendTable = func(t *Table) bool {
		for _, rt := range t.Refs {
			if _, ok := done[rt]; !ok {
				return false
			}
		}
		tbls = append(tbls, t)
		done[t.Name] = struct{}{}
		for _, pt := range pendings[t.Name] {
			appendTable(pt)
		}
		delete(pendings, t.Name)
		return true
	}

	for _, n := range names {
		var t Table
		err := db.Get(&t, fmt.Sprintf("SHOW CREATE TABLE `%s`", n))
		if err != nil {
			return nil, err
		}
		for _, ref := range reRef.FindAllStringSubmatch(t.Create, -1) {
			t.Refs = append(t.Refs, ref[1])
		}
		if !appendTable(&t) {
			for _, rt := range t.Refs {
				if _, ok := done[rt]; !ok {
					pendings[rt] = append(pendings[rt], &t)
				}
			}
		}
	}

	if len(tbls) != len(names) {
		var p []string
		for _, n := range names {
			if _, ok := done[n]; !ok {
				p = append(p, n)
			}
		}
		return nil, fmt.Errorf("pending tables: %v", strings.Join(p, ", "))
	}

	return tbls, nil
}

// GetProcedures returns stored procedure informatins
func GetProcedures(db *sqlx.DB) ([]*Procedure, error) {
	// Note: go-mysql-server does not currently support "SHOW PROCEDURE STATUS" or "information_schema.routines".
	// This function returns only the required stored procedures.
	// Update this implementation if support is added in the future.
	var nms []string
	nms = append(nms, "_migration_exists")

	procs := make([]*Procedure, 0, len(nms))
	for _, n := range nms {
		var p Procedure
		err := db.Get(&p, fmt.Sprintf("SHOW CREATE PROCEDURE `%s`", n))
		if err != nil {
			return nil, err
		}
		procs = append(procs, &p)
	}

	return procs, nil
}
