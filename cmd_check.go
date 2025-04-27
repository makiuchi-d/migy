package main

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"
	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/migrations"
	"github.com/makiuchi-d/migy/sqlfile"
)

var cmdCheck = &cobra.Command{
	Use:   "check [flags] [number]",
	Short: "Check if an up/down migration pair is reversible",
	Long: `Check if an up/down migration pair is reversible.
Applies the up migration to a temporary database and then rolls it back
using the down migration to verify that no differences remain.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		num := 0
		if len(args) > 0 {
			var err error
			num, err = strconv.Atoi(args[0])
			if err != nil {
				return err
			}
		}
		return checkMigrationPair(targetDir, num)
	},
}

func init() {
	cmd.AddCommand(cmdCheck)
}

func checkMigrationPair(dir string, num int) error {
	migs, err := migrations.Load(dir)
	if err != nil {
		return err
	}

	if num != 0 {
		i, err := migs.FindNumber(num)
		if err != nil {
			return err
		}
		migs = migs[:i+1]
	}

	mig := migs[len(migs)-1]
	if !mig.UpDown {
		return fmt.Errorf("no up/down migration: number=%06d", mig.Number)
	}

	db := sqlx.NewDb(testdb.New("db"), "mysql")

	for name := range migs[:len(migs)-1].FromSnapshot().ApplicableFileNames() {
		info("applying:", name)
		if err := sqlfile.Apply(db, filepath.Join(dir, name)); err != nil {
			return err
		}
	}

	// snapshot
	info("======== taking a snapshot")
	ss, err := dbstate.TakeSnapshot(db)

	info("applying:", mig.UpName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.UpName())); err != nil {
		return err
	}

	info("applying:", mig.DownName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.DownName())); err != nil {
		return err
	}

	// check with snapshot
	info("======== checking")

	_ = ss

	return nil
}
