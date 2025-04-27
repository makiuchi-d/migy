package dbstate_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/makiuchi-d/migy/dbstate"
)

func TestRow_String(t *testing.T) {
	var n, s, d any
	n = 42
	d, _ = time.Parse(time.DateTime, "2025-04-26 19:22:30")
	s = "line1\r\n\t'%'"
	row := dbstate.Row{&n, &d, &s}
	exp := "(42, '2025-04-26 19:22:30', 'line1\\r\\n\\t\\'\\%\\'')"

	if r := row.String(); r != exp {
		t.Fatalf("Row: %q\nwants %q", r, exp)
	}
}

func TestGetRecords(t *testing.T) {
	db := prepareTestDb(t)

	rec, err := dbstate.GetRecords(db, "users")
	if err != nil {
		t.Fatal(err)
	}

	expcol := []string{"id", "name"}
	exprows := []string{"(1, 'alice')", "(2, 'bob')", "(3, 'carol')"}

	if !reflect.DeepEqual(rec.Columns, expcol) {
		t.Errorf("colmuns: %v wants %v", rec.Columns, expcol)
	}

	for i, row := range rec.Rows {
		if rs, es := row.String(), exprows[i]; rs != es {
			t.Errorf("row[%v]: %q wants %q", i, rs, es)
		}
	}
}
