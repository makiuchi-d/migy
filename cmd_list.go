package main

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/migrations"
)

var cmdList = &cobra.Command{
	Use:   "list [flags] [DUMP_FILE | --host HOST DB_NAME | --dsn DSN]",
	Short: "List migration files needed to reach the target state",
	Long: `Lists migration files needed to reach the target migration number by
comparing the migration directory with the database or dump file.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDBorDumpfile(args)
		if err != nil {
			return err
		}
		if db == nil {
			return errors.New("data source or dump file is required")
		}

		return printFilesToApply(db, targetDir, targetNum)
	},
}

func init() {
	cmd.AddCommand(cmdList)
	addFlagNumber(cmdList)
	addFlagsForDB(cmdList)
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
		num = migs.Last().Number
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
