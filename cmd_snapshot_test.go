package main

import (
	"fmt"
	"io" // For io.Copy
	"os"
	"path/filepath"

	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"

	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/sqlfile"
)

func copyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	return err
}

func copyFiles(src, dst string, files ...string) error {
	for _, f := range files {
		err := copyFile(filepath.Join(src, f), filepath.Join(dst, f))
		if err != nil {
			return err
		}
	}
	return nil
}

func TestSnapshotToSQLFile(t *testing.T) {
	ignores := map[string][]string{"_migrations": {"applied"}}

	dir := t.TempDir()

	if err := copyFiles(filepath.Join("testdata", "snapshot"), dir,
		"000000_init.all.sql",
		"000010_create_users.up.sql",
		"000010_create_users.down.sql",
		"000020_alter_users.up.sql",
		"000020_alter_users.down.sql",
	); err != nil {
		t.Fatalf("copy migration files: %v", err)
	}

	// expected db snapshots
	var ss10, ss20 *dbstate.Snapshot
	err := func() error {
		db := sqlx.NewDb(testdb.New("db"), "mysql")
		defer db.Close()

		err := sqlfile.Apply(db, filepath.Join(dir, "000000_init.all.sql"))
		if err != nil {
			return fmt.Errorf("apply init sql: %w", err)
		}
		err = sqlfile.Apply(db, filepath.Join(dir, "000010_create_users.up.sql"))
		if err != nil {
			return fmt.Errorf("apply sql(10): %w", err)
		}
		ss10, err = dbstate.TakeSnapshot(db)
		if err != nil {
			return fmt.Errorf(": %w", err)
		}

		err = sqlfile.Apply(db, filepath.Join(dir, "000020_alter_users.up.sql"))
		if err != nil {
			return fmt.Errorf("apply sql(20): %w", err)
		}
		ss20, err = dbstate.TakeSnapshot(db)
		if err != nil {
			return fmt.Errorf(": %w", err)
		}

		return nil
	}()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		num  int
		file string
		exp  *dbstate.Snapshot
	}{
		"snapshot10": {10, "000010_create_users.all.sql", ss10},
		"snapshot20": {-1, "000020_alter_users.all.sql", ss20},
	}
	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			if err := snapshotToSQLFile(dir, test.num, false); err != nil {
				t.Fatalf("snapshotToSQLFile: %v", err)
			}

			db := sqlx.NewDb(testdb.New("db"), "mysql")
			defer db.Close()

			if err := sqlfile.Apply(db, filepath.Join(dir, test.file)); err != nil {
				t.Fatalf("apply snapshot: %v", err)
			}
			diff, err := dbstate.Diff(db, test.exp, ignores)
			if err != nil {
				t.Fatalf("diff snapshot: %v", err)
			}
			if diff != "" {
				t.Fatal("diff snapshot:\n" + diff)
			}
		})
	}
}
