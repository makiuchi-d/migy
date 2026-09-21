package dbstate

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Snapshot struct {
	Tables    map[string]*Table
	Records   map[string]*Records
	Sequences map[string]*Sequence
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

	seqs, err := GetSequences(db)
	if err != nil {
		return nil, err
	}
	seqmap := make(map[string]*Sequence, len(seqs))
	for i := range seqs {
		seqmap[seqs[i].Name] = seqs[i]
	}

	return &Snapshot{
		Tables:    tblmap,
		Records:   recmap,
		Sequences: seqmap,
	}, nil
}
