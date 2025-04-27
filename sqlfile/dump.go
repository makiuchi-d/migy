package sqlfile

import (
	"fmt"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/makiuchi-d/migy/dbstate"
)

func Dump(w io.Writer, db *sqlx.DB) error {
	schema, err := dbstate.GetSchema(db)
	if err != nil {
		return err
	}

	// tables
	for _, t := range schema.Tables {
		w.Write([]byte(t.Create))
		w.Write([]byte(";\n\n"))

		rec, err := dbstate.GetRecords(db, t.Name)
		if err != nil {
			return err
		}
		if len(rec.Rows) == 0 {
			break
		}

		for i, r := range rec.Rows {
			if i%10 == 0 {
				fmt.Fprintf(w, "INSERT INTO `%v` (`%v`) VALUES\n  ", t.Name, strings.Join(rec.Columns, "`,`"))
			}

			w.Write([]byte(r.String()))

			if i%10 == 10-1 || i == len(rec.Rows)-1 {
				w.Write([]byte(";\n"))
			} else {
				w.Write([]byte(", "))
			}
		}
		w.Write([]byte("\n"))
	}

	// stored procedures
	w.Write([]byte("DELIMITER //\n\n"))
	for _, p := range schema.Procedures {
		w.Write([]byte(p.Create))
		w.Write([]byte("//\n\n"))
	}
	w.Write([]byte("DELIMITER ;\n"))

	return nil
}
