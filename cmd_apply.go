package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/sqlfile"
)

var cmdApply = &cobra.Command{
	Use:   "apply [flags] [--host HOST DB_NAME | --dsn DSN]",
	Short: "Apply migration files to reach the target database state",
	Long: `Apply migration files to the target database in order.
Continues from the last applied migration and moves forward or backward
to match the target migration number.
This command requires a live database connection.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := openDB(args)
		if err != nil {
			return err
		}

		confirm := func(msg func()) bool {
			if applyYes {
				return true
			}
			msg()
			fmt.Print("Do you want to continue? [y/N]: ")
			s := bufio.NewScanner(os.Stdin)
			s.Scan()
			in := strings.ToLower(strings.TrimSpace(s.Text()))
			return in == "y" || in != "yes"
		}

		ok, err := applyMigrations(db, targetDir, targetNum, confirm)
		if err != nil {
			return err
		}
		if !ok {
			os.Exit(1)
		}
		return nil
	},
}

var applyYes bool

func init() {
	cmd.AddCommand(cmdApply)
	addFlagNumber(cmdApply)
	addFlagsForDB(cmdApply)
	cmdApply.Flags().BoolVarP(&applyYes, "yes", "y", false, "assume \"yes\" as answer to all prompts")
}

func applyMigrations(db *sqlx.DB, dir string, num int, confirm func(func()) bool) (bool, error) {
	files, err := listFilesToApply(db, dir, num)
	if err != nil {
		return false, err
	}
	if len(files) == 0 {
		info("Nothing to do.")
		return true, nil
	}
	abort := !confirm(func() {
		info("The following migration files will be applied:")
		for _, file := range files {
			info(" -", file)
		}
	})
	if abort {
		info("Abort.")
		return false, nil
	}

	for _, file := range files {
		info("applying:", file)
		if err := sqlfile.Apply(db, filepath.Join(dir, file)); err != nil {
			return false, err
		}
	}

	return true, nil
}
