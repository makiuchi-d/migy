package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
		if checkFrom >= 0 {
			return checkMigrationsFrom(targetDir, checkFrom, targetNum)
		}

		diff, err := checkMigration(targetDir, targetNum)
		if err != nil {
			return err
		}

		if diff != "" {
			info(diff, "\ncheck failed")
			os.Exit(1)
		}

		return nil
	},
}

var checkFrom int

func init() {
	cmd.AddCommand(cmdCheck)
	addFlagNumber(cmdCheck)
	cmdCheck.Flags().IntVarP(&checkFrom, "from", "", -1, "check each migration from this to --number")
	cmdCheck.Flags().Lookup("from").DefValue = "n"
}

// chheckMigrationsFrom checks migrations from specified number step by step.
func checkMigrationsFrom(dir string, from, to int) error {
	if to > 0 && from > to {
		return fmt.Errorf("--from (%06d) must be less than %06d", from, to)
	}
	migs, err := migrations.Load(dir)
	f, err := migs.FindNumber(from)
	if err != nil {
		return err
	}
	migs = migs[f:]
	if to >= 0 {
		t, err := migs.FindNumber(to)
		if err != nil {
			return err
		}
		migs = migs[:t+1]
	}

	// use a subprocess per check to isolate in-process db and release memory after each run.
	cmd := []string{
		os.Args[0], // [0]
		"check",    // [1]
		"-n",       // [2]
		"",         // [3]: placeholder
	}
	if dir != "." {
		cmd = append(cmd, "-d", dir)
	}
	if quit {
		cmd = append(cmd, "-q")
	}

	for _, mig := range migs {
		info(fmt.Sprintf("==== check %06d", mig.Number))
		cmd[3] = strconv.Itoa(mig.Number)
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			os.Exit(c.ProcessState.ExitCode())
		}
	}

	return nil
}

func checkMigration(dir string, num int) (string, error) {
	migs, err := migrations.Load(dir)
	if err != nil {
		return "", err
	}

	if num >= 0 {
		i, err := migs.FindNumber(num)
		if err != nil {
			return "", err
		}
		migs = migs[:i+1]
	}

	if len(migs) < 2 {
		return "", fmt.Errorf("no migration to check")
	}

	mig := migs.Last()
	if !mig.UpDown {
		return "", fmt.Errorf("no up/down migration: number=%06d", mig.Number)
	}

	files, err := migs[:len(migs)-1].FileNamesFromSnapshot()
	if err != nil {
		return "", err
	}

	db := sqlx.NewDb(testdb.New("db"), "mysql")
	defer db.Close()

	for _, file := range files {
		info("applying:", file)
		if err := sqlfile.Apply(db, filepath.Join(dir, file)); err != nil {
			return "", err
		}
	}

	// snapshot for up/down check
	ss, err := dbstate.TakeSnapshot(db)
	if err != nil {
		return "", err
	}

	info("---- up/down")
	info("applying:", mig.UpName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.UpName())); err != nil {
		return "", err
	}

	// snapshot for .all.sql check
	var ss2 *dbstate.Snapshot
	if mig.Snapshot {
		ss2, err = dbstate.TakeSnapshot(db)
		if err != nil {
			return "", err
		}
	}

	info("applying:", mig.DownName())
	if err := sqlfile.Apply(db, filepath.Join(dir, mig.DownName())); err != nil {
		return "", err
	}

	info("checking...")
	diff, err := dbstate.Diff(db, ss, mig.Ignores)
	if err != nil {
		return "", err
	}
	if diff != "" {
		return strings.TrimSuffix(diff, "\n"), nil
	}
	info("ok")
	if !mig.Snapshot {
		return "", nil
	}

	info("---- snapshot")
	db2 := sqlx.NewDb(testdb.New("db2"), "mysql")
	defer db2.Close()
	info("applying:", mig.SnapshotName())
	if err := sqlfile.Apply(db2, filepath.Join(dir, mig.SnapshotName())); err != nil {
		return "", err
	}
	info("checking...")
	diff, err = dbstate.Diff(db2, ss2, map[string][]string{"_migrations": {"applied"}})
	if err != nil {
		return "", err
	}
	if diff != "" {
		return strings.TrimSuffix(diff, "\n"), nil
	}
	info("ok")

	return "", nil
}
