package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"

	"golang.org/x/exp/maps"
)

var (
	ErrNoMigration     = errors.New("no migration found")
	ErrDuplicateNumber = errors.New("duplicate number")
	ErrMissingFile     = errors.New("missing")
	ErrTitleMismatch   = errors.New("title mismatch")

	namePattern *regexp.Regexp = regexp.MustCompile(`^([0-9]+)_(.*)\.(up|down|all)\.sql$`)
)

func parseSQLFileName(name string) (num int, title, kind string, ok bool) {
	m := namePattern.FindStringSubmatch(name)
	if len(m) != 4 {
		return 0, "", "", false
	}
	n, _ := strconv.Atoi(m[1])
	return n, m[2], m[3], true
}

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
	if !m.UpDown {
		return ""
	}
	return fmt.Sprintf("%06d_%s.up.sql", m.Number, m.Title)
}

// DownName returns filename of '*.down.sql'.
func (m *Migration) DownName() string {
	if !m.UpDown {
		return ""
	}
	return fmt.Sprintf("%06d_%s.down.sql", m.Number, m.Title)
}

// SnapshotName returns filename of '*.all.sql' if exists
func (m *Migration) SnapshotName() string {
	if !m.Snapshot {
		return ""
	}
	return fmt.Sprintf("%06d_%s.all.sql", m.Number, m.Title)
}

// GetMigrations returns all migration SQL files in the dir.
// This list is sorted by its number.
func GetMigrations(dir string) (Migrations, error) {
	dent, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	mm := make(map[int]struct {
		title string
		kinds map[string]string
	})

	for _, ent := range dent {
		if ent.IsDir() {
			continue
		}

		name := ent.Name()
		n, title, kind, ok := parseSQLFileName(name)
		if !ok {
			continue
		}

		m, ok := mm[n]
		if !ok {
			m.title = title
			m.kinds = make(map[string]string)
			m.kinds[kind] = name
			mm[n] = m
			continue
		}

		if nm, ok := m.kinds[kind]; ok {
			return nil, fmt.Errorf("%w: %s, %s", ErrDuplicateNumber, nm, name)
		}
		if title != m.title {
			return nil, fmt.Errorf("%w: %06d %q, %q", ErrTitleMismatch, n, m.title, title)
		}

		m.kinds[kind] = name
	}

	if len(mm) == 0 {
		return nil, ErrNoMigration
	}

	keys := maps.Keys(mm)
	slices.Sort(keys)
	migs := make([]*Migration, len(keys))
	for i, num := range keys {
		m := mm[num]

		_, up := m.kinds["up"]
		_, down := m.kinds["down"]
		_, all := m.kinds["all"]

		if !up && down {
			return nil, fmt.Errorf("%w up.sql: %06d", ErrMissingFile, num)
		}
		if up && !down {
			return nil, fmt.Errorf("%w down.sql: %06d", ErrMissingFile, num)
		}

		migs[i] = &Migration{
			Number:   num,
			Title:    m.title,
			UpDown:   up,
			Snapshot: all,
		}
	}

	return migs, nil
}

func (migs Migrations) Newest() *Migration {
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].UpDown {
			return migs[i]
		}
	}
	return nil
}

func (migs Migrations) FromSnapshot() Migrations {
	if len(migs) == 0 {
		return migs //empty
	}
	return migs.FromSnapshotTo(migs[len(migs)-1].Number)
}

func (migs Migrations) FromSnapshotTo(num int) Migrations {
	last := -1
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].Number == num {
			last = i
			break
		}
	}
	if last < 0 {
		return Migrations{}
	}
	for i := last - 1; i >= 0; i-- {
		if migs[i].Snapshot {
			return migs[i : last+1]
		}
	}
	warning("no snapshot (*.all.sql) found")
	return migs[:last+1]
}
