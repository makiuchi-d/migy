package migrations

import (
	"errors"
	"fmt"
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

// FromSnapshot returns migrations from the latest snapshot to the latest migration
func (migs Migrations) FromSnapshot() Migrations {
	if len(migs) == 0 {
		return migs //empty
	}
	migs, _ = migs.FromSnapshotTo(migs[len(migs)-1].Number)
	return migs
}

// FromSnapshotTo returns migrations from the snapshot before the specified migration to the specified migration
func (migs Migrations) FromSnapshotTo(num int) (Migrations, error) {
	last := -1
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].Number == num {
			last = i
			break
		}
	}
	if last < 0 {
		return nil, fmt.Errorf("%w: number=%06d", ErrNoMigration, num)
	}
	for i := last - 1; i >= 0; i-- {
		if migs[i].Snapshot {
			return migs[i : last+1], nil
		}
	}
	return migs[:last+1], nil
}
