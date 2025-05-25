package main

import (
	"fmt"
	"path/filepath"

	"github.com/jmoiron/sqlx"
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

		return listUpgradeFiles(db, targetDir, migNumber)
	},
}

func init() {
	cmd.AddCommand(cmdList)
	cmdList.Flags().IntVarP(&migNumber, "number", "n", 0, "migration number")
}

func listUpgradeFiles(db *sqlx.DB, dir string, num int) error {
	hists, err := migrations.LoadHistories(db)
	if err != nil {
		return err
	}
	cur := 0
	if len(hists) > 0 {
		cur = hists[len(hists)-1].Id
	}

	migs, err := migrations.Load(dir)
	if err != nil {
		return err
	}
	if num == 0 && len(migs) > 0 {
		num = migs[len(migs)-1].Number
	}

	files, err := migs.FilenamesToApply(cur, num)
	if err != nil {
		return err
	}

	for f := range files {
		fmt.Println(filepath.Join(dir, f))
	}

	return nil
}
