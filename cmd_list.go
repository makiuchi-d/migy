package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/migrations"
	"github.com/spf13/cobra"
)

var cmdList = &cobra.Command{
	Use:   "list [flags] <dsn|dumpfile>",
	Short: "List unapplied migration files",
	Long: `Lists unapplied migration files by comparing the migration directory
with the given database or dump file.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("missing dsn or dumpfile")
		}

		db, err := openDsnOrDumpfile(args[0])
		if err != nil {
			return err
		}

		return printFilesToApply(db, targetDir, migNumber)
	},
}

func init() {
	cmd.AddCommand(cmdList)
	addFlagNumber(cmdList)
}

func printFilesToApply(db *sqlx.DB, dir string, num int) error {
	files, err := listFilesToApply(db, dir, num)
	if err != nil {
		return err
	}
	for _, file := range files {
		fmt.Println(filepath.Join(dir, file))
	}
	return nil
}

func listFilesToApply(db *sqlx.DB, dir string, num int) ([]string, error) {
	migs, err := migrations.Load(dir)
	if err != nil {
		return nil, err
	}
	if num < 0 {
		num = migs[len(migs)-1].Number
	}

	err = dbstate.HasMigrationTable(db)
	if err != nil {
		if !errors.Is(err, dbstate.ErrNoMigrationTable) {
			return nil, err
		}

		// show files from snapshot
		i, err := migs.FindNumber(num)
		if err != nil {
			return nil, err
		}
		return migs[:i+1].FileNamesFromSnapshot()
	}

	hists, err := migrations.LoadHistories(db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("'_migrations' table found but not initialized")
		}
		return nil, err
	}
	cur := hists.CurrentNum()

	ms := make(migrations.Migrations, 0, len(migs))
	for s := range migrations.BuildStatus(migs, hists) {
		ms = append(ms, s.Migration)
	}

	return ms.FileNamesToApply(cur, num)
}
