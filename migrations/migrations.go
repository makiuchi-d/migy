package migrations

import (
	"errors"
	"fmt"
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
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].Number == num {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%w: number=%06d", ErrNoMigration, num)
}

// FindNext returns the index of the next migration for the specified number.
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
func (migs Migrations) FindLatestSnapshot() (int, error) {
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].Snapshot {
			return i, nil
		}
	}
	return -1, fmt.Errorf("%w: no snapshot (*.all.sql))", ErrNoMigration)
}

// FileNamesFromSnapshot returns the sql files to restore db state
func (migs Migrations) FileNamesFromSnapshot() ([]string, error) {
	st, err := migs.FindLatestSnapshot()
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(migs)-st)
	files = append(files, migs[st].SnapshotName())

	for _, m := range migs[st+1:] {
		if !m.UpDown {
			return nil, fmt.Errorf("%w: number=%06d", ErrSequenceGap, m.Number)
		}
		files = append(files, m.UpName())
	}
	return files, nil
}

// FileNamesToApply returns the sql files to upgrade or downgrade from the current number to the target number
func (migs Migrations) FileNamesToApply(current, target int) ([]string, error) {
	if current == target {
		return nil, nil
	}

	t, err := migs.FindNumber(target)
	if err != nil {
		return nil, err
	}

	if current < target {
		// upgrade
		c, err := migs[:t+1].FindNext(current)
		if err != nil {
			return nil, err
		}
		files := make([]string, 0, t-c+1)
		for _, m := range migs[c : t+1] {
			if !m.UpDown {
				return nil, fmt.Errorf("%w: number=%06d", ErrSequenceGap, m.Number)
			}
			files = append(files, m.UpName())
		}
		return files, nil
	} else {
		// downgrade
		c, err := migs.FindNumber(current)
		if err != nil {
			return nil, err
		}
		files := make([]string, 0, c-t+1)
		for i := c; i > t; i-- {
			m := migs[i]
			if !m.UpDown {
				return nil, fmt.Errorf("%w: number=%06d", ErrSequenceGap, m.Number)
			}
			files = append(files, m.DownName())
		}
		return files, nil
	}
}
