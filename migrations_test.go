package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMigrations(t *testing.T) {
	migs, err := GetMigrations("testdata/migrations")
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
			_, err := GetMigrations("testdata/" + dir)
			t.Logf("%v", err)
			if !errors.Is(err, exp) {
				t.Errorf("%v wants %v", err, exp)
			}
		})
	}
}

func TestNewestMigration(t *testing.T) {
	migs, err := GetMigrations("testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}
	mig := migs.Newest()
	if mig.Number != 4 {
		t.Fatalf("migration is not newest: %v", mig.Number)
	}
}

func TestFromSnapshot(t *testing.T) {
	migs, err := GetMigrations("testdata/migrations")
	if err != nil {
		t.Fatal(err)
	}

	exp := []int{3, 4}
	nums := make([]int, 0, len(migs))
	for _, m := range migs.FromSnapshot() {
		nums = append(nums, m.Number)
	}
	if !reflect.DeepEqual(nums, exp) {
		t.Fatalf("from snapshot: %v wants %v", nums, exp)
	}

	exp = []int{0, 1, 3}
	nums = nums[:0]
	for _, m := range migs.FromSnapshotTo(3) {
		nums = append(nums, m.Number)
	}
	if !reflect.DeepEqual(nums, exp) {
		t.Fatalf("from snapshot to '3': %v wants %v", nums, exp)
	}
}
