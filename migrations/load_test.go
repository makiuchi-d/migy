package migrations

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLoad(t *testing.T) {
	migs, err := Load("testdata/migrations")
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}

	emptymap := make(map[string][]string)

	exp := Migrations{
		{
			Number:   0,
			Title:    "init",
			UpDown:   false,
			Snapshot: true,
			Ignores:  emptymap,
		},
		{
			Number:   1,
			Title:    "foo",
			UpDown:   true,
			Snapshot: false,
			Ignores:  emptymap,
		},
		{
			Number:   3,
			Title:    "bar",
			UpDown:   true,
			Snapshot: true,
			Ignores:  emptymap,
		},
		{
			Number:   4,
			Title:    "baz",
			UpDown:   true,
			Snapshot: false,
			// migy:ignore mytable.ignore1 mytable.ignore2
			Ignores: map[string][]string{
				"mytable": {"ignore1", "ignore2"},
			},
		},
	}

	if diff := cmp.Diff(migs, exp); diff != "" {
		t.Fatal(diff)
	}
}

func TestLoadFail(t *testing.T) {
	tests := map[string]error{
		"empty":       ErrNoMigration,
		"duplicate":   ErrDuplicateNumber,
		"mismatch":    ErrTitleMismatch,
		"upmissing":   ErrMissingFile,
		"downmissing": ErrMissingFile,
		"invalidform": ErrInvalidFormat,
	}
	for dir, exp := range tests {
		t.Run(dir, func(t *testing.T) {
			_, err := Load("testdata/" + dir)
			if !errors.Is(err, exp) {
				t.Errorf("%v wants %v", err, exp)
			}
		})
	}
}

func TestReadIgnores(t *testing.T) {
	igs := make(map[string][]string)
	err := readIgnores(igs, "testdata/ignore/000001_ignore.down.sql")
	if err != nil {
		t.Fatal(err)
	}

	exp := map[string][]string{
		"table1": {"column1", "column2"},
		"table2": {"*"},
	}

	if diff := cmp.Diff(igs, exp); diff != "" {
		t.Fatal(diff)
	}
}
