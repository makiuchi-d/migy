package sqlfile_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"

	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/sqlfile"
)

func strarr[A any](arr []A, f func(a A) string) []string {
	s := make([]string, 0, len(arr))
	for i := range arr {
		s = append(s, f(arr[i]))
	}
	return s
}

func TestApply(t *testing.T) {
	db := sqlx.NewDb(testdb.New("db"), "mysql")

	err := sqlfile.Apply(db, "testdata/apply/apply.sql")
	if err != nil {
		t.Fatal(err)
	}

	tables, err := dbstate.GetTables(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) != 1 || tables[0].Name != "memo" {
		t.Fatalf("unexpected tables: %q",
			strarr(tables, func(t *dbstate.Table) string { return t.Name }))
	}

	recs, err := dbstate.GetRecords(db, "memo")
	if err != nil {
		t.Fatal(err)
	}

	records := []string{
		"(1, 'first', '2025-04-27 10:00:00')",
		"(2, 'secnd', '2025-04-27 11:00:00')",
		"(3, 'third', '2025-04-27 12:00:00')",
	}
	diff := cmp.Diff(
		strarr(recs.Rows, func(r dbstate.Row) string { return r.String() }),
		records)
	if diff != "" {
		t.Fatal(diff)
	}
}
