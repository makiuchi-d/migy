package dbstate_test

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/maps"

	"github.com/makiuchi-d/migy/dbstate"
)

func TestTakeSnapshot(t *testing.T) {
	db := prepareTestDb(t)

	ss, err := dbstate.TakeSnapshot(db)
	if err != nil {
		t.Fatal(err)
	}

	exptbls := []string{"_migrations", "users", "user_emails"}
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
		"user_emails": {
			Columns: []string{"user_id", "email"},
			Rows:    []string{"(1, 'alice@example.com')", "(2, 'bob1@example.com')", "(2, 'bob2@example.com')"},
		},
	}

	if len(ss.Tables) != len(exptbls) {
		t.Errorf("tables: %v wants %v", maps.Keys(ss.Tables), exptbls)
	} else {
		for _, tbl := range ss.Tables {
			if !slices.Contains(exptbls, tbl.Name) {
				t.Errorf("unexpected table: %v", tbl.Name)
			}
		}
	}

	if len(ss.Records) != len(exprecs) {
		t.Fatalf("all records: %v wants %v", len(ss.Records), len(exprecs))
	} else {
		for name, rec := range ss.Records {
			exp, ok := exprecs[name]
			if !ok {
				t.Errorf("unexpected table records: %v", name)
				continue
			}

			if diff := cmp.Diff(rec.Columns, exp.Columns); diff != "" {
				t.Errorf("columns of %v:\n%v", name, diff)
				continue
			}

			var rows []string
			for _, r := range rec.Rows {
				rows = append(rows, r.String())
			}
			diff := cmp.Diff(rows, exp.Rows)
			if diff != "" {
				t.Errorf("rows of %v:\n%v", name, diff)
			}
		}
	}

	if len(ss.Procedures) != len(expprocs) {
		t.Errorf("procedures: %v wants %v", maps.Keys(ss.Procedures), expprocs)
	} else {
		for _, proc := range ss.Procedures {
			if !slices.Contains(expprocs, proc.Name) {
				t.Errorf("unexpected procedure: %v", proc.Name)
			}
		}
	}
}
