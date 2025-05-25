package migrations

import (
	"errors"
	"fmt"
	"iter"
)

var (
	ErrNoMigration     = errors.New("no migration found")
	ErrDuplicateNumber = errors.New("duplicate number")
	ErrMissingFile     = errors.New("missing")
	ErrTitleMismatch   = errors.New("title mismatch")
	ErrInvalidFormat   = errors.New("invalid format")
	ErrSequenceGap     = errors.New("sequence gap")
)

// Migration SQL file info
type Migration struct {
	Number   int
	Title    string
	UpDown   bool
	Snapshot bool
	Ignores  map[string][]string
}

// Migration list
type Migrations []*Migration

// UpName returns filename of '*.up.sql'.
func (m *Migration) UpName() string {
	return fmt.Sprintf("%06d_%s.up.sql", m.Number, m.Title)
}

// DownName returns filename of '*.down.sql'.
func (m *Migration) DownName() string {
	return fmt.Sprintf("%06d_%s.down.sql", m.Number, m.Title)
}

// SnapshotName returns filename of '*.all.sql' if exists
func (m *Migration) SnapshotName() string {
	return fmt.Sprintf("%06d_%s.all.sql", m.Number, m.Title)
}

// FindNumber returns the index of the migration with the specified number
func (migs Migrations) FindNumber(num int) (int, error) {
	i := len(migs) - 1
	for ; i >= 0; i-- {
		if migs[i].Number == num {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%w: number=%06d", ErrNoMigration, num)
}

func (migs Migrations) FindNext(num int) (int, error) {
	idx := -1
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].Number > num {
			idx = i
		} else {
			break
		}
	}
	if idx < 0 {
		return -1, fmt.Errorf("%w: number > %06d", ErrNoMigration, num)
	}
	return idx, nil
}

// FindLatestSnapshot returns the index of the migration with snapshot
func (migs Migrations) FindLatestSnapshot() int {
	i := len(migs) - 1
	for ; i >= 0; i-- {
		if migs[i].Snapshot {
			return i
		}
	}
	return i
}

// FromSnapshot returns migrations from the latest snapshot to the latest migration
func (migs Migrations) FromSnapshot() Migrations {
	if len(migs) == 0 {
		return migs //empty
	}
	i := migs.FindLatestSnapshot()
	return migs[max(i, 0):]
}

// FromSnapshotTo returns migrations from the snapshot before the specified migration to the specified migration
func (migs Migrations) FromSnapshotTo(num int) (Migrations, error) {
	last, err := migs.FindNumber(num)
	if err != nil {
		return nil, err
	}

	start := migs[:last].FindLatestSnapshot()
	return migs[max(start, 0) : last+1], nil
}

func (migs Migrations) ApplicableFileNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		if len(migs) == 0 {
			return
		}
		f := migs[0].SnapshotName()
		if !migs[0].Snapshot {
			f = migs[0].UpName()
		}
		if !yield(f) {
			return
		}

		for _, mig := range migs[1:] {
			if !mig.UpDown {
				continue
			}
			if !yield(mig.UpName()) {
				return
			}
		}
	}
}

func (migs Migrations) FilenamesToApply(current, target int) (iter.Seq[string], error) {
	if current == target {
		return func(_ func(string) bool) {}, nil
	}

	reverse := false
	if current > target {
		reverse = true
		current, target = target, current
	}

	start, err := migs.FindNext(current)
	if err != nil {
		return nil, err
	}
	last, err := migs.FindNumber(target)
	if err != nil {
		return nil, err
	}

	migs = migs[start : last+1]

	for _, m := range migs {
		if !m.UpDown {
			return nil, fmt.Errorf("%w: number=%06d", ErrSequenceGap, m.Number)
		}
	}

	if !reverse {
		return func(yield func(string) bool) {
			for _, m := range migs {
				if !yield(m.UpName()) {
					return
				}
			}
		}, nil
	} else {
		return func(yield func(string) bool) {
			for i := len(migs) - 1; i >= 0; i-- {
				if !yield(migs[i].DownName()) {
					return
				}
			}
		}, nil
	}
}
