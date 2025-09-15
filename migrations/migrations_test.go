package migrations

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLast(t *testing.T) {
	migs := Migrations{
		{Number: 0},
		{Number: 10},
		{Number: 20},
	}

	m := migs.Last()
	if m.Number != 20 {
		t.Errorf("Last: Number=%v, wants 20", m.Number)
	}
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
		err           error
	}{
		"0-5": {0, 5, 5, nil},
		"0-4": {0, 4, 3, nil},
		"0-3": {0, 3, 3, nil},
		"0-2": {0, 2, 0, nil},
		"1-2": {1, 2, -1, ErrNoMigration},
	}

	for k, test := range tests {
		mm := migs[test.from : test.to+1]
		i, err := mm.FindLatestSnapshot()
		if !errors.Is(err, test.err) {
			t.Errorf("%v: error: %v wants %v", k, err, test.err)
			continue
		}
		if i != test.exp {
			t.Errorf("%v: %v (%06d) wants %v (%06d)",
				k, i, mm[i].Number, test.exp, mm[test.exp].Number)
		}
	}
}

func TestFileNamesFromSnapshot(t *testing.T) {
	migs := Migrations{
		{Number: 0, Title: "init", Snapshot: true},
		{Number: 10, Title: "first", UpDown: true, Snapshot: false},
		{Number: 20, Title: "second", UpDown: false, Snapshot: true},
		{Number: 30, Title: "third", UpDown: true, Snapshot: false},
		{Number: 40, Title: "fourth", UpDown: true, Snapshot: false},
	}

	_, err := migs[3:].FileNamesFromSnapshot()
	if !errors.Is(err, ErrNoMigration) {
		t.Errorf("must be ErrNoMigration: %v", err)
	}

	tests := map[string]struct {
		migs Migrations
		exp  []string
	}{
		"[:2]": {migs[:2], []string{
			"000000_init.all.sql",
			"000010_first.up.sql",
		}},
		"[2:3]": {migs[2:3], []string{
			"000020_second.all.sql",
		}},
		"[:]": {migs[:], []string{
			"000020_second.all.sql",
			"000030_third.up.sql",
			"000040_fourth.up.sql",
		}},
	}
	for k, test := range tests {
		files, err := test.migs.FileNamesFromSnapshot()
		if err != nil {
			t.Errorf("%v error: %v", k, err)
			continue
		}
		if diff := cmp.Diff(test.exp, files); diff != "" {
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
		files, err := migs.FileNamesToApply(test.current, test.target)
		if !errors.Is(err, test.err) {
			t.Errorf("%v: error=%q wants %q", k, err, test.err)
			continue
		}

		if diff := cmp.Diff(test.exp, files); diff != "" {
			t.Errorf("%q: %v", k, diff)
		}
	}
}
