package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"
	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/migrations"
	"github.com/makiuchi-d/migy/sqlfile"
)

var cmdSnapshot = &cobra.Command{
	Use:   "snapshot",
	Short: "Generate a SQL snapshot at the specified migration point",
	Long: `Applies up migrations sequentially and generates a SQL file
that reproduces the database state at that point.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return snapshotToSQLFile(targetDir, targetNum, overwrite)
	},
}

func init() {
	cmd.AddCommand(cmdSnapshot)
	addFlagNumber(cmdSnapshot)
	addFlagForce(cmdSnapshot)
}

func snapshotToSQLFile(dir string, num int, overwrite bool) error {

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
		return fmt.Errorf("no migration to make a snapshot")
	}

	mig := migs.Last()
	if !mig.UpDown {
		return fmt.Errorf("no up/down migration: number=%06d", mig.Number)
	}
	if mig.Snapshot && !overwrite {
		return fmt.Errorf("file exists: %s", filepath.Join(dir, mig.SnapshotName()))
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
	info("applying:", mig.UpName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.UpName())); err != nil {
		return err
	}

	info("========\nwriting:", mig.SnapshotName())
	flag := os.O_CREATE | os.O_RDWR | os.O_TRUNC
	f, err := os.OpenFile(filepath.Join(dir, mig.SnapshotName()), flag, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprint(f, signature, "\n\n"); err != nil {
		return err
	}
	return sqlfile.Dump(f, db)
}
