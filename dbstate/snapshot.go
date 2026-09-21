package dbstate

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Snapshot struct {
	Tables  map[string]*Table
	Records map[string]*Records
	Objects map[string]*Object
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

	objs, err := GetObjects(db)
	if err != nil {
		return nil, err
	}
	objmap := make(map[string]*Object, len(objs))
	for i := range objs {
		objmap[objs[i].Name] = objs[i]
	}

	return &Snapshot{
		Tables:  tblmap,
		Records: recmap,
		Objects: objmap,
	}, nil
}
