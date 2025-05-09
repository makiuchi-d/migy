package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

var (
	reFilenname = regexp.MustCompile(`^([0-9]+)_(.*)\.(up|down|all)\.sql$`)
	reIgnore    = regexp.MustCompile(`\smigy:ignore\s+(.*)+?(?:\n|$)`)
)

func parseSQLFileName(name string) (num int, title, kind string, ok bool) {
	m := reFilenname.FindStringSubmatch(name)
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

		_, all := m.kinds["all"]
		_, up := m.kinds["up"]
		downname, down := m.kinds["down"]

		if !up && down {
			return nil, fmt.Errorf("%w up.sql: %06d", ErrMissingFile, num)
		}
		if up && !down {
			return nil, fmt.Errorf("%w down.sql: %06d", ErrMissingFile, num)
		}

		ignores := make(map[string][]string)
		if down {
			err := readIgnores(ignores, filepath.Join(dir, downname))
			if err != nil {
				return nil, fmt.Errorf("%v: %w", downname, err)
			}
		}

		migs[i] = &Migration{
			Number:   num,
			Title:    m.title,
			UpDown:   up,
			Snapshot: all,
			Ignores:  ignores,
		}
	}

	return migs, nil
}

func readIgnores(igs map[string][]string, name string) error {
	file, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	m := reIgnore.FindAllStringSubmatch(string(file), -1)

	for _, s := range m {
		for ss := range strings.FieldsSeq(s[1]) {
			tc := strings.Split(strings.TrimSuffix(ss, ","), ".")
			if len(tc) != 2 {
				return fmt.Errorf("%w: %v", ErrInvalidFormat, s[0])
			}
			igs[tc[0]] = append(igs[tc[0]], tc[1])
		}
	}

	return nil
}
