package sqlfile

import (
	"fmt"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/makiuchi-d/migy/dbstate"
)

func Dump(w io.Writer, db *sqlx.DB) error {
	// Sequences come first: a table may refer to one from a column DEFAULT,
	// whether it owns the sequence or not.
	// Identity sequences are not listed here, the column definition creates them.
	seqs, err := dbstate.GetSequences(db)
	if err != nil {
		return err
	}
	owned := make(map[string][]*dbstate.Sequence)
	for _, s := range seqs {
		fmt.Fprintf(w, "%s;\n", s.Create)
		if s.OwnerTable.Valid {
			owned[s.OwnerTable.String] = append(owned[s.OwnerTable.String], s)
		}
	}
	if len(seqs) > 0 {
		fmt.Fprint(w, "\n")
	}

	tbls, err := dbstate.GetTables(db)
	if err != nil {
		return err
	}
	for _, t := range tbls {
		fmt.Fprintf(w, "%s;\n\n", t.Create)

		for _, s := range owned[t.Name] {
			fmt.Fprintf(w, "%s;\n", s.OwnedBy())
		}
		for _, idx := range t.Indexes {
			fmt.Fprintf(w, "%s;\n", idx)
		}
		if len(owned[t.Name])+len(t.Indexes) > 0 {
			fmt.Fprint(w, "\n")
		}

		if err := dumpRecords(w, db, t); err != nil {
			return err
		}
	}

	// The rows above were inserted with explicit values, so every counter is
	// still where the sequence was created.
	vals, err := dbstate.GetSequenceValues(db)
	if err != nil {
		return err
	}
	for _, v := range vals {
		fmt.Fprintf(w, "SELECT setval(%v, %v, %v);\n",
			quoteLiteral(quoteIdent(v.Name)), v.Last, v.IsCalled)
	}

	return nil
}

func dumpRecords(w io.Writer, db *sqlx.DB, t *dbstate.Table) error {
	name := quoteIdent(t.Name)

	rec, err := dbstate.GetRecords(db, t.Name)
	if err != nil {
		return err
	}
	if len(rec.Rows) == 0 {
		return nil
	}

	// A generated column cannot be given a value, an identity column
	// declared GENERATED ALWAYS needs OVERRIDING SYSTEM VALUE.
	colmap := make(map[string]*dbstate.Column, len(t.Columns))
	for _, c := range t.Columns {
		colmap[strings.Trim(c.Name, `"`)] = c
	}
	cols := make([]string, 0, len(rec.Columns))
	keep := make([]int, 0, len(rec.Columns))
	overriding := ""
	for i, n := range rec.Columns {
		c, ok := colmap[n]
		if ok && c.Generated {
			continue
		}
		if ok && c.Identity == "a" {
			overriding = "OVERRIDING SYSTEM VALUE "
		}
		cols = append(cols, quoteIdent(n))
		keep = append(keep, i)
	}

	for i, r := range rec.Rows {
		if i%10 == 0 {
			fmt.Fprintf(w, "INSERT INTO %v (%v) %vVALUES\n  ",
				name, strings.Join(cols, ", "), overriding)
		}

		row := make(dbstate.Row, 0, len(keep))
		for _, k := range keep {
			row = append(row, r[k])
		}
		fmt.Fprint(w, row.String())

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

func quoteLiteral(s string) string {
	return `'` + strings.ReplaceAll(s, `'`, `''`) + `'`
}
