package sqlfile

import (
	"fmt"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/makiuchi-d/migy/dbstate"
)

func Dump(w io.Writer, db *sqlx.DB) error {
	// Objects that belong to no table come first: a table may refer to one.
	objs, err := dbstate.GetObjects(db)
	if err != nil {
		return err
	}
	for _, o := range objs {
		fmt.Fprintf(w, "%s\n\n", o.Create)
	}

	tbls, err := dbstate.GetTables(db)
	if err != nil {
		return err
	}
	for _, t := range tbls {
		fmt.Fprintf(w, "%s\n\n", t.Create)

		if err := dumpRecords(w, db, t); err != nil {
			return err
		}
	}

	return nil
}

func dumpRecords(w io.Writer, db *sqlx.DB, t *dbstate.Table) error {
	rec, err := dbstate.GetRecords(db, t.Name)
	if err != nil {
		return err
	}
	if len(rec.Rows) == 0 {
		return nil
	}

	cols := make([]string, len(rec.Columns))
	for i, c := range rec.Columns {
		cols[i] = quoteIdent(c)
	}

	name := quoteIdent(t.Name)
	for i, r := range rec.Rows {
		if i%10 == 0 {
			fmt.Fprintf(w, "INSERT INTO %v (%v) %vVALUES\n  ",
				name, strings.Join(cols, ", "), rec.InsertOption)
		}

		fmt.Fprint(w, r.String())

		if i%10 == 10-1 || i == len(rec.Rows)-1 {
			fmt.Fprint(w, ";\n")
		} else {
			fmt.Fprint(w, ", ")
		}
	}
	fmt.Fprint(w, "\n")

	return nil
}

func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
