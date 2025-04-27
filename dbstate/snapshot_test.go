package dbstate_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/makiuchi-d/migy/dbstate"
)

func strarr[A any](arr []A, f func(a A) string) []string {
	s := make([]string, 0, len(arr))
	for i := range arr {
		s = append(s, f(arr[i]))
	}
	return s
}

func TestGetAllRecords(t *testing.T) {
	db := prepareTestDb(t)

	ss, err := dbstate.TakeSnapshot(db)
	if err != nil {
		t.Fatal(err)
	}

	exptbls := []string{"_migrations", "users"}
	expprocs := []string{"_migration_exists"}
	exprecs := map[string]struct {
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

	if len(ss.Tables) != len(exptbls) {
		t.Errorf("tables: %v wants %v",
			strarr(ss.Tables, func(t *dbstate.Table) string { return t.Name }),
			exptbls)
	} else {
		for _, tbl := range ss.Tables {
			if !slices.Contains(exptbls, tbl.Name) {
				t.Errorf("unexpected table: %v", tbl.Name)
			}
		}
	}

	if len(ss.AllRecords) != len(exprecs) {
		t.Fatalf("all records: %v wants %v", len(ss.AllRecords), len(exprecs))
	} else {
		for name, rec := range ss.AllRecords {
			exp, ok := exprecs[name]
			if !ok {
				t.Errorf("unexpected table records: %v", name)
				continue
			}

			if diff := cmp.Diff(rec.Columns, exp.Columns); diff != "" {
				t.Errorf("columns of %v:\n%v", name, diff)
				continue
			}

			diff := cmp.Diff(strarr(rec.Rows, func(r dbstate.Row) string { return r.String() }), exp.Rows)
			if diff != "" {
				t.Errorf("rows of %v:\n%v", name, diff)
			}
		}
	}

	if len(ss.Procedures) != len(expprocs) {
		t.Errorf("procedures: %v wants %v",
			strarr(ss.Procedures, func(p *dbstate.Procedure) string { return p.Name }),
			expprocs)
	} else {
		for _, proc := range ss.Procedures {
			if !slices.Contains(expprocs, proc.Name) {
				t.Errorf("unexpected procedure: %v", proc.Name)
			}
		}
	}
}
