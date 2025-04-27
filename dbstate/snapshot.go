package dbstate

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Snapshot struct {
	Tables     []*Table
	AllRecords map[string]*Records
	Procedures []*Procedure
}

func TakeSnapshot(db *sqlx.DB) (*Snapshot, error) {
	tbls, err := GetTables(db)
	if err != nil {
		return nil, err
	}
	recs, err := getAllRecords(db, tbls)
	if err != nil {
		return nil, err
	}
	procs, err := GetProcedures(db)
	if err != nil {
		return nil, err
	}

	return &Snapshot{
		Tables:     tbls,
		AllRecords: recs,
		Procedures: procs,
	}, nil
}

func getAllRecords(db *sqlx.DB, tbls []*Table) (map[string]*Records, error) {
	recs := make(map[string]*Records, len(tbls))
	for _, tbl := range tbls {
		rec, err := GetRecords(db, tbl.Name)
		if err != nil {
			return nil, fmt.Errorf("table %s: %w", tbl.Name, err)
		}
		recs[tbl.Name] = rec
	}
	return recs, nil
}
