package migrations

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindNumber(t *testing.T) {
	migs := Migrations{
		{Number: 0, Snapshot: true},
		{Number: 10, UpDown: true, Snapshot: false},
		{Number: 20, UpDown: true, Snapshot: false},
		{Number: 30, UpDown: true, Snapshot: true},
		{Number: 40, UpDown: true, Snapshot: false},
		{Number: 50, UpDown: false, Snapshot: true},
	}

	tests := map[int]struct {
		num int
		err error
	}{
		0:  {0, nil},
		30: {3, nil},
		50: {5, nil},
		15: {-1, ErrNoMigration},
	}
	for num, exp := range tests {
		i, err := migs.FindNumber(num)
		if exp.num != i || (err == nil) != (exp.err == nil) || !errors.Is(err, exp.err) {
			t.Errorf("num=%v: index=%v (%v) wants %v (%v)", num, i, err, exp.num, exp.err)
		}
	}
}

func TestFindLatestSnapshot(t *testing.T) {
	migs := Migrations{
		{Number: 0, Snapshot: true},
		{Number: 10, UpDown: true, Snapshot: false},
		{Number: 20, UpDown: true, Snapshot: false},
		{Number: 30, UpDown: true, Snapshot: true},
		{Number: 40, UpDown: true, Snapshot: false},
		{Number: 50, UpDown: false, Snapshot: true},
	}

	tests := map[string]struct {
		from, to, exp int
	}{
		"0-5": {0, 5, 5},
		"0-4": {0, 4, 3},
		"0-3": {0, 3, 3},
		"0-2": {0, 2, 0},
		"1-2": {1, 2, -1},
	}

	for k, test := range tests {
		mm := migs[test.from : test.to+1]
		i := mm.FindLatestSnapshot()
		if i != test.exp {
			t.Errorf("%v: %v (%06d) wants %v (%06d)",
				k, i, mm[i].Number, test.exp, mm[test.exp].Number)
		}
	}
}

func TestFromSnapshot(t *testing.T) {
	migs, err := Load("testdata/migrations")
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
	mi, err := migs.FromSnapshotTo(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range mi {
		nums = append(nums, m.Number)
	}
	if !reflect.DeepEqual(nums, exp) {
		t.Fatalf("from snapshot to '3': %v wants %v", nums, exp)
	}
}

func TestApplicableFileNames(t *testing.T) {
	migs := Migrations{
		{Number: 0, Title: "init", Snapshot: true},
		{Number: 10, Title: "first", UpDown: true, Snapshot: false},
		{Number: 20, Title: "second", UpDown: false, Snapshot: true},
		{Number: 30, Title: "third", UpDown: true, Snapshot: true},
	}
	exp := []string{
		"000000_init.all.sql",
		"000010_first.up.sql",
		"000030_third.up.sql",
	}

	var names []string
	for n := range migs.ApplicableFileNames() {
		names = append(names, n)
	}
	if diff := cmp.Diff(names, exp); diff != "" {
		t.Fatal(diff)
	}

	names = names[:0]
	for n := range migs[1:].ApplicableFileNames() {
		names = append(names, n)
	}
	if diff := cmp.Diff(names, exp[1:]); diff != "" {
		t.Fatal(diff)
	}
}
