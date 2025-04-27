package migrations

import (
	"errors"
	"fmt"
	"iter"
	"regexp"
)

var (
	ErrNoMigration     = errors.New("no migration found")
	ErrDuplicateNumber = errors.New("duplicate number")
	ErrMissingFile     = errors.New("missing")
	ErrTitleMismatch   = errors.New("title mismatch")

	namePattern *regexp.Regexp = regexp.MustCompile(`^([0-9]+)_(.*)\.(up|down|all)\.sql$`)
)

// Migration SQL file info
type Migration struct {
	Number   int
	Title    string
	UpDown   bool
	Snapshot bool
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
