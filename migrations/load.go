package migrations

import (
	"fmt"
	"os"
	"slices"
	"strconv"

	"golang.org/x/exp/maps"
)

func parseSQLFileName(name string) (num int, title, kind string, ok bool) {
	m := namePattern.FindStringSubmatch(name)
	if len(m) != 4 {
		return 0, "", "", false
	}
	n, _ := strconv.Atoi(m[1])
	return n, m[2], m[3], true
}

// Load returns all migration SQL files in the dir.
// This list is sorted by its number.
func Load(dir string) (Migrations, error) {
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
