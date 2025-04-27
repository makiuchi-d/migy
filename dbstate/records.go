package dbstate

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Records struct {
	Columns []string
	Rows    []Row
}

type Row []any

func (r Row) String() string {
	if len(r) == 0 {
		return "()"
	}
	b := []byte{'('}
	for _, col := range r {
		switch v := (*col.(*any)).(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			b = fmt.Appendf(b, "%v, ", v)
		case time.Time:
			b = v.AppendFormat(b, "'2006-01-02 15:04:05', ")
		default:
			b = append(b, []byte(quotedValue(v))...)
			b = append(b, ',', ' ')
		}
	}
	b[len(b)-2] = ')'
	return string(b[:len(b)-1])
}

func quotedValue(v any) string {
	s := fmt.Sprintf("%v", v)
	var b strings.Builder
	b.Grow(len(s) + 2)

	b.WriteByte('\'')
	for _, c := range s {
		switch c {
		case 0:
			b.WriteString("\\0")
		case 26: // ^Z, SUB
			b.WriteString("\\Z")
		case '\b':
			b.WriteString("\\b")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		case '\'', '\\', '%', '_':
			b.WriteByte('\\')
			b.WriteRune(c)
		default:
			b.WriteRune(c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

func GetRecords(db *sqlx.DB, table string) (*Records, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	rec := Records{
		Columns: cols,
	}

	for rows.Next() {
		row := make([]any, len(cols))
		for i := range len(cols) {
			row[i] = new(any)
		}
		if err := rows.Scan(row...); err != nil {
			return nil, err
		}
		rec.Rows = append(rec.Rows, row)
	}

	return &rec, nil
}

func GetAllRecords(db *sqlx.DB, schema *Schema) (map[string]*Records, error) {
	recs := make(map[string]*Records, len(schema.Tables))
	for _, tbl := range schema.Tables {
		rec, err := GetRecords(db, tbl.Name)
		if err != nil {
			return nil, fmt.Errorf("table %s: %w", tbl.Name, err)
		}
		recs[tbl.Name] = rec
	}
	return recs, nil
}
