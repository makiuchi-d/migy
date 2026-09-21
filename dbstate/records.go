package dbstate

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Records struct {
	Columns []string
	Rows    []Row
}

// Row holds the values of a record, already rendered as SQL literals
// by quote_nullable(): a NULL is the bare word NULL.
type Row []string

func (r Row) String() string {
	return "(" + strings.Join(r, ", ") + ")"
}

const qRecordColumns = `
SELECT a.attname
  FROM pg_attribute a
 WHERE a.attrelid = to_regclass($1) AND a.attnum > 0 AND NOT a.attisdropped
 ORDER BY a.attnum`

const qPrimaryKey = `
SELECT quote_ident(a.attname)
  FROM pg_index i
  JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
 WHERE i.indrelid = to_regclass($1) AND i.indisprimary
 ORDER BY array_position(i.indkey::int2[], a.attnum)`

// GetRecords returns all records of the table as SQL literals.
//
// Rendering is left to the server: quote_nullable() applies the output
// function of each type, so json, bytea, arrays and the like come back in a
// form that can be fed straight back into an INSERT, with no guessing from
// whatever Go type the driver happened to choose.
//
// Rows are ordered by the primary key. The heap order of PostgreSQL is not
// stable (an UPDATE moves the row), which would show up as a false difference.
func GetRecords(db *sqlx.DB, table string) (*Records, error) {
	name := quoteIdent(table)

	var cols []string
	if err := db.Select(&cols, qRecordColumns, name); err != nil {
		return nil, err
	}

	var pk []string
	if err := db.Select(&pk, qPrimaryKey, name); err != nil {
		return nil, err
	}

	rec := Records{Columns: cols}
	if len(cols) == 0 {
		return &rec, nil
	}

	quoted := make([]string, len(cols))
	for i, c := range cols {
		quoted[i] = fmt.Sprintf("quote_nullable(%s)", quoteIdent(c))
	}

	q := fmt.Sprintf("SELECT %s FROM %s", strings.Join(quoted, ", "), name)
	if len(pk) > 0 {
		q += " ORDER BY " + strings.Join(pk, ", ")
	}

	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		row := make(Row, len(cols))
		ptrs := make([]any, len(cols))
		for i := range row {
			ptrs[i] = &row[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		rec.Rows = append(rec.Rows, row)
	}

	return &rec, rows.Err()
}
