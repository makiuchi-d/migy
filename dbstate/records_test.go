package dbstate_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

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

func TestGetAllRecords(t *testing.T) {
	db := prepareTestDb(t)

	schema, err := dbstate.GetSchema(db)
	if err != nil {
		t.Fatal(err)
	}

	recs, err := dbstate.GetAllRecords(db, schema)
	if err != nil {
		t.Fatal(err)
	}

	exps := map[string]struct {
		Columns []string
		Rows    []string
	}{
		"_migrations": {
			Columns: []string{"id", "applied", "title"},
			Rows:    []string{"(1, '2025-04-19 00:33:32', 'first')"},
		},
		"users": {
			Columns: []string{"id", "name"},
			Rows:    []string{"(1, 'alice')", "(2, 'bob')", "(3, 'carol')"},
		},
	}

	if len(recs) != len(exps) {
		t.Fatalf("count of recs: %v wants %v", len(recs), len(exps))
	}
	for k, rec := range recs {
		exp, ok := exps[k]
		if !ok {
			t.Fatalf("unexpected table: %v", k)
		}

		if diff := cmp.Diff(rec.Columns, exp.Columns); diff != "" {
			t.Fatalf("table %v:\n%v", k, diff)
		}

		if len(rec.Rows) != len(exp.Rows) {
			t.Fatalf("table %v: count of rows: %v wants %v", k, len(rec.Rows), len(exp.Rows))
		}

		for i := range rec.Rows {
			if r, e := rec.Rows[i].String(), exp.Rows[i]; r != e {
				t.Fatalf("table %v row[%v]: %v wants %v", k, i, r, e)
			}
		}
	}
}
