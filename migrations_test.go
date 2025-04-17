package main

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMigrations(t *testing.T) {
	migs, err := Migrations("testdata/migrations")
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}

	exp := []*Migration{
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

func TestMigrationsFail(t *testing.T) {
	tests := map[string]error{
		"empty":       ErrNoMigration,
		"duplicate":   ErrDuplicateNumber,
		"mismatch":    ErrTitleMismatch,
		"upmissing":   ErrMissingFile,
		"downmissing": ErrMissingFile,
	}
	for dir, exp := range tests {
		t.Run(dir, func(t *testing.T) {
			_, err := Migrations("testdata/" + dir)
			t.Logf("%v", err)
			if !errors.Is(err, exp) {
				t.Errorf("%v wants %v", err, exp)
			}
		})
	}
}

func TestNewestMigration(t *testing.T) {
	mig, err := NewestMigration("testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}
	if mig.Number != 4 {
		t.Fatalf("migration is not newest: %v", mig.Number)
	}
}

func TestNewestSnapshot(t *testing.T) {
	mig, err := NewestSnapshot("testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}
	if mig.Number != 3 {
		t.Fatalf("snapshot is not newest: %v", mig.Number)
	}
}
