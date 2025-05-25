package migrations

import (
	"errors"
	"iter"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func all[T any](s iter.Seq[T]) []T {
	var a []T
	for v := range s {
		a = append(a, v)
	}
	return a
}

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

func TestFindNext(t *testing.T) {
	migs := Migrations{
		{Number: 0, Snapshot: true},
		{Number: 10, UpDown: true, Snapshot: false},
		{Number: 20, UpDown: true, Snapshot: false},
		{Number: 30, UpDown: true, Snapshot: true},
		{Number: 40, UpDown: true, Snapshot: false},
		{Number: 50, UpDown: false, Snapshot: true},
	}

	tests := map[int]struct {
		idx int
		err error
	}{
		0:  {1, nil},
		29: {3, nil},
		30: {4, nil},
		31: {4, nil},
		50: {-1, ErrNoMigration},
	}
	for num, exp := range tests {
		i, err := migs.FindNext(num)
		if exp.err != nil && !errors.Is(err, exp.err) {
			t.Errorf("num=%v: err=%q wants %q", num, err, exp.err)
			continue
		}

		if i != exp.idx {
			t.Errorf("num=%v: index=%v (%v) wants %v (%v)",
				num, i, migs[i].Number, exp.idx, migs[exp.idx].Number)
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
		{Number: 30, Title: "third", UpDown: true, Snapshot: false},
		{Number: 40, Title: "fourth", UpDown: true, Snapshot: false},
	}

	_, err := migs.ApplicableFileNames()
	if err == nil || !errors.Is(err, ErrSequenceGap) {
		t.Errorf("must be ErrSequenceGap: %v", err)
	}

	tests := map[string]struct {
		migs Migrations
		exp  []string
	}{
		"[:2]": {migs[:2], []string{
			"000000_init.all.sql",
			"000010_first.up.sql",
		}},
		"[2:]": {migs[2:], []string{
			"000020_second.all.sql",
			"000030_third.up.sql",
			"000040_fourth.up.sql",
		}},
		"[3:]": {migs[3:], []string{
			"000030_third.up.sql",
			"000040_fourth.up.sql",
		}},
	}
	for k, test := range tests {
		f, err := test.migs.ApplicableFileNames()
		if err != nil {
			t.Errorf("%v error: %v", k, err)
			continue
		}
		if diff := cmp.Diff(all(f), test.exp); diff != "" {
			t.Fatalf("%v: %v", k, diff)
		}
	}
}

func TestFileNamesToApply(t *testing.T) {
	migs := Migrations{
		{Number: 0, Title: "init", Snapshot: true},
		{Number: 10, Title: "first", UpDown: true, Snapshot: false},
		{Number: 20, Title: "second", UpDown: true, Snapshot: true},
		{Number: 30, Title: "third", UpDown: false, Snapshot: true},
		{Number: 40, Title: "fourth", UpDown: true, Snapshot: false},
		{Number: 50, Title: "fifth", UpDown: true, Snapshot: true},
	}

	tests := map[string]struct {
		current int
		target  int
		exp     []string
		err     error
	}{
		"0 to 20":  {0, 20, []string{"000010_first.up.sql", "000020_second.up.sql"}, nil},
		"20 to 0":  {20, 0, []string{"000020_second.down.sql", "000010_first.down.sql"}, nil},
		"10 to 10": {10, 10, nil, nil},
		"10 to 40": {10, 40, nil, ErrSequenceGap},
		"30 to 50": {30, 50, []string{"000040_fourth.up.sql", "000050_fifth.up.sql"}, nil},
		"50 to 30": {50, 30, []string{"000050_fifth.down.sql", "000040_fourth.down.sql"}, nil},
	}

	for k, test := range tests {
		f, err := migs.FileNamesToApply(test.current, test.target)
		if test.err != nil {
			if !errors.Is(err, test.err) {
				t.Errorf("%v: error=%q wants %q", k, err, test.err)
			}
			continue
		}

		if diff := cmp.Diff(all(f), test.exp); diff != "" {
			t.Errorf("%q: %v", k, diff)
		}
	}
}
