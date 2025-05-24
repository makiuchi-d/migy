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

var (
	snapshotNumber    int
	snapshotOverwrite bool
)

var cmdSnapshot = &cobra.Command{
	Use:   "snapshot",
	Short: "Generate a SQL snapshot at the specified migration point",
	Long: `Applies up migrations sequentially and generates a SQL file
that reproduces the database state at that point.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return snapshotToSQLFile(targetDir, snapshotNumber, snapshotOverwrite)
	},
}

func init() {
	cmd.AddCommand(cmdSnapshot)
	cmdSnapshot.Flags().IntVarP(&snapshotNumber, "number", "n", 0, "migration number")
	cmdSnapshot.Flags().BoolVarP(&snapshotOverwrite, "force", "f", false, "Override the output file if it exists")
}

func snapshotToSQLFile(dir string, num int, overwrite bool) error {

	migs, err := migrations.Load(dir)
	if err != nil {
		return err
	}

	if num == 0 {
		num = migs[len(migs)-1].Number
	}

	migs, err = migs.FromSnapshotTo(num)
	if err != nil {
		return err
	}
	if !migs[0].Snapshot {
		warning("no snapshot (*.all.sql)")
	}

	last := migs[len(migs)-1]
	if last.Snapshot && !overwrite {
		return fmt.Errorf("file exists: %s", filepath.Join(dir, last.SnapshotName()))
	}

	db := sqlx.NewDb(testdb.New("db"), "mysql")

	for name := range migs.ApplicableFileNames() {
		info("applying:", name)
		if err := sqlfile.Apply(db, filepath.Join(dir, name)); err != nil {
			return err
		}
	}

	info("========\nwriting:", last.SnapshotName())
	flag := os.O_CREATE | os.O_RDWR | os.O_TRUNC
	f, err := os.OpenFile(filepath.Join(dir, last.SnapshotName()), flag, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprint(f, signature, "\n\n"); err != nil {
		return err
	}
	return sqlfile.Dump(f, db)
}
