package main

import (
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/migy/sqlfile"
	"github.com/spf13/cobra"
)

var cmdApply = &cobra.Command{
	Use:   "apply [flags] [--host HOST DB_NAME | --dsn DSN]",
	Short: "Apply unapplied up migration files to the database",
	Long: `Apply unapplied up migration files to the target database in order.
Continues from the last applied migration and applies the remaining files in sequence.
This command requires a live database connection.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB(args)
		if err != nil {
			return err
		}

		return applyMigrations(db, targetDir, migNumber)
	},
}

func init() {
	cmd.AddCommand(cmdApply)
	addFlagNumber(cmdApply)
	addFlagsForDB(cmdApply)
}

func applyMigrations(db *sqlx.DB, dir string, num int) error {
	files, err := listFilesToApply(db, dir, num)
	if err != nil {
		return err
	}
	for _, file := range files {
		sqlfile.Apply(db, filepath.Join(dir, file))
	}
	return nil
}
