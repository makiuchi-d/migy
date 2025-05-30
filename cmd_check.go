package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"
	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/migrations"
	"github.com/makiuchi-d/migy/sqlfile"
)

var cmdCheck = &cobra.Command{
	Use:   "check [flags]",
	Short: "Check if an up/down migration pair is reversible",
	Long: `Check if an up/down migration pair is reversible.
Applies the up migration to a temporary database and then rolls it back
using the down migration to verify that no differences remain.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return checkMigrationPair(targetDir, migNumber)
	},
}

func init() {
	cmd.AddCommand(cmdCheck)
	addFlagNumber(cmdCheck)
}

func checkMigrationPair(dir string, num int) error {
	migs, err := migrations.Load(dir)
	if err != nil {
		return err
	}

	if num >= 0 {
		i, err := migs.FindNumber(num)
		if err != nil {
			return err
		}
		migs = migs[:i+1]
	}

	if len(migs) < 2 {
		return fmt.Errorf("no migration to check")
	}

	mig := migs[len(migs)-1]
	if !mig.UpDown {
		return fmt.Errorf("no up/down migration: number=%06d", mig.Number)
	}

	files, err := migs[:len(migs)-1].FileNamesFromSnapshot()
	if err != nil {
		return err
	}

	db := sqlx.NewDb(testdb.New("db"), "mysql")

	for _, file := range files {
		info("applying:", file)
		if err := sqlfile.Apply(db, filepath.Join(dir, file)); err != nil {
			return err
		}
	}

	// snapshot for up/down check
	ss, err := dbstate.TakeSnapshot(db)
	if err != nil {
		return err
	}

	info("---- up/down")
	info("applying:", mig.UpName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.UpName())); err != nil {
		return err
	}

	// snapshot for .all.sql check
	var ss2 *dbstate.Snapshot
	if mig.Snapshot {
		ss2, err = dbstate.TakeSnapshot(db)
		if err != nil {
			return err
		}
	}

	info("applying:", mig.DownName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.DownName())); err != nil {
		return err
	}

	info("checking...")
	diff, err := dbstate.Diff(db, ss, mig.Ignores)
	if err != nil {
		return err
	}
	if diff != "" {
		info(strings.TrimSuffix(diff, "\n"), "\ncheck failed")
		os.Exit(1)
	}
	info("ok")
	if !mig.Snapshot {
		return nil
	}

	info("---- snapshot")
	db2 := sqlx.NewDb(testdb.New("db2"), "mysql")
	info("applying:", mig.SnapshotName())
	if err := sqlfile.Apply(db2, filepath.Join(dir, mig.SnapshotName())); err != nil {
		return err
	}
	info("checking...")
	diff, err = dbstate.Diff(db2, ss2, map[string][]string{"_migrations": {"applied"}})
	if err != nil {
		return err
	}
	if diff != "" {
		info(strings.TrimSuffix(diff, "\n"), "\ncheck failed")
		os.Exit(1)
	}
	info("ok")

	return nil
}
