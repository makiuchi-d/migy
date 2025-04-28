package dbstate

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Snapshot struct {
	Tables     map[string]*Table
	Records    map[string]*Records
	Procedures map[string]*Procedure
}

func TakeSnapshot(db *sqlx.DB) (*Snapshot, error) {
	tbls, err := GetTables(db)
	if err != nil {
		return nil, err
	}
	tblmap := make(map[string]*Table, len(tbls))
	recmap := make(map[string]*Records, len(tbls))
	for i := range tbls {
		n := tbls[i].Name
		tblmap[n] = tbls[i]
		rec, err := GetRecords(db, n)
		if err != nil {
			return nil, fmt.Errorf("table %s: %w", n, err)
		}
		recmap[n] = rec
	}

	procs, err := GetProcedures(db)
	if err != nil {
		return nil, err
	}
	procmap := make(map[string]*Procedure)
	for i := range procs {
		n := procs[i].Name
		procmap[n] = procs[i]
	}

	return &Snapshot{
		Tables:     tblmap,
		Records:    recmap,
		Procedures: procmap,
	}, nil
}
