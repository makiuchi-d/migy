package migrations_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/makiuchi-d/migy/migrations"
)

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
}
