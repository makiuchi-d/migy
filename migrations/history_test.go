package migrations_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"

	"github.com/makiuchi-d/migy/migrations"
)

func TestLoadHistories(t *testing.T) {
	create := "CREATE TABLE _migrations (" +
		"id      INTEGER NOT NULL," +
		"applied DATETIME," +
		"title   VARCHAR(255)," +
		"PRIMARY KEY (id))"
	insert := "INSERT INTO _migrations (id, applied, title) VALUES" +
		"(0, '2025-09-15 10:20:30', 'init')," +
		"(10, '2025-09-15 11:22:33', 'first')"

	exp := migrations.Histories{
		{0, time.Date(2025, 9, 15, 10, 20, 30, 0, time.UTC), "init"},
		{10, time.Date(2025, 9, 15, 11, 22, 33, 0, time.UTC), "first"},
	}

	db := sqlx.NewDb(testdb.New("db"), "mysql")
	if _, err := db.Exec(create); err != nil {
		t.Fatalf("db.Exec(create): %v", err)
	}
	if _, err := db.Exec(insert); err != nil {
		t.Fatalf("db.Exec(insert): %v", err)
	}

	hs, err := migrations.LoadHistories(db)
	if err != nil {
		t.Fatalf("LoadHistories: %v", err)
	}

	if diff := cmp.Diff(hs, exp); diff != "" {
		t.Fatal(diff)
	}
}

func TestCurrentNum(t *testing.T) {
	hists := migrations.Histories{
		{1, time.Date(2025, time.May, 10, 1, 4, 7, 0, time.Local), "first"},
		{3, time.Date(2025, time.May, 11, 2, 5, 8, 0, time.Local), "third"},
		{4, time.Date(2025, time.May, 12, 3, 6, 9, 0, time.Local), "fourth-db"},
	}

	exp := 4
	if n := hists.CurrentNum(); n != exp {
		t.Errorf("CurrentNum: %v, wants %v", n, exp)
	}

	hists = hists[:2]
	exp = 3
	if n := hists.CurrentNum(); n != exp {
		t.Errorf("CurrentNum: %v, wants %v", n, exp)
	}

	hists = hists[:0]
	exp = -1
	if n := hists.CurrentNum(); n != exp {
		t.Errorf("CurrentNum: %v, wants %v", n, exp)
	}
}

func TestBuildHistories(t *testing.T) {
	dt1 := time.Date(2025, time.May, 10, 1, 4, 7, 0, time.Local)
	dt3 := time.Date(2025, time.May, 11, 2, 5, 8, 0, time.Local)
	dt4 := time.Date(2025, time.May, 12, 3, 6, 9, 0, time.Local)
	hists := []migrations.History{
		{1, dt1, "first"},
		{3, dt3, "third"},
		{4, dt4, "fourth-db"},
	}
	migs := []*migrations.Migration{
		{0, "init", false, true, nil},
		{1, "first", true, false, nil},
		{2, "second", true, true, nil},
		{4, "fourth", true, false, nil},
		{5, "fifth", true, false, nil},
	}
	exp := []migrations.Status{
		{&migrations.Migration{0, "init", false, true, nil}, time.Time{}, ""},
		{&migrations.Migration{1, "first", true, false, nil}, dt1, ""},
		{&migrations.Migration{2, "second", true, true, nil}, time.Time{}, ""},
		{&migrations.Migration{3, "third", false, false, nil}, dt3, ""},
		{&migrations.Migration{4, "fourth", true, false, nil}, dt4, "fourth-db"},
		{&migrations.Migration{5, "fifth", true, false, nil}, time.Time{}, ""},
	}

	var ss []migrations.Status
	for s := range migrations.BuildStatus(migs, hists) {
		ss = append(ss, s)
	}

	if diff := cmp.Diff(ss, exp); diff != "" {
		t.Fatal(diff)
	}

	if a, exp := ss[5].IsApplied(), false; a != exp {
		t.Errorf("statses[5].IsAPplied: %v, wants %v", a, exp)
	}
	if a, exp := ss[4].IsApplied(), true; a != exp {
		t.Errorf("statses[5].IsAPplied: %v, wants %v", a, exp)
	}
}
