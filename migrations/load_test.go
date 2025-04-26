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

	exp := Migrations{
		{
			Number:   0,
			Title:    "init",
			UpDown:   false,
			Snapshot: true,
		},
		{
			Number:   1,
			Title:    "foo",
			UpDown:   true,
			Snapshot: false,
		},
		{
			Number:   3,
			Title:    "bar",
			UpDown:   true,
			Snapshot: true,
		},
		{
			Number:   4,
			Title:    "baz",
			UpDown:   true,
			Snapshot: false,
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
