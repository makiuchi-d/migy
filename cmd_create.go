package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/makiuchi-d/migy/migrations"
)

// cmdCreate represents the create command
var cmdCreate = &cobra.Command{
	Use:   "create [flags] <TITLE>",
	Short: "Create a new pair of up/down SQL migration files",
	Long: `Create a new pair of up/down SQL migration files.
The up file defines the forward migration. The down file contains
the corresponding rollback, ensuring changes can be reversed.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("title is required")
		}
		return createNewMigrationFiles(targetDir, targetNum, args[0])
	},
}

func init() {
	cmd.AddCommand(cmdCreate)
	addFlagNumber(cmdCreate)
}

const (
	createUpSQL = signature + `
INSERT INTO _migrations (id, title, applied) VALUES ({{.Number}}, '{{.Title}}', now());
-- Write your forward migration SQL statements below.
`

	createDownSQL = signature + `
CALL _migration_exists({{.Number}});
DELETE FROM _migrations WHERE id = {{.Number}};
-- Write your rollback SQL statements below.
`
)

func nextNum(n int) int {
	return n + 10 - n%10
}

func createNewMigrationFiles(dir string, num int, title string) error {
	migs, err := migrations.Load(dir)
	if err != nil {
		return err
	}

	if num < 0 {
		num = nextNum(migs.Last().Number)
	} else {
		for i := len(migs) - 1; i >= 0; i-- {
			n := migs[i].Number
			if n < num {
				break
			}
			if n == num {
				return fmt.Errorf("duplicate number: %06d %v", n, migs[i].Title)
			}
		}
	}

	mig := migrations.Migration{
		Number: num,
		Title:  title,
	}

	err = generateMigrationSQLFile(dir, mig.UpName(), createUpSQL, mig)
	if err != nil {
		return fmt.Errorf("up sql file (%s,%s): %w", dir, mig.UpName(), err)
	}

	err = generateMigrationSQLFile(dir, mig.DownName(), createDownSQL, mig)
	if err != nil {
		return fmt.Errorf("down sql file: %w", err)
	}

	return nil
}

func generateMigrationSQLFile(dir, name, tmpl string, mig migrations.Migration) error {
	f, err := os.OpenFile(filepath.Join(dir, name), os.O_CREATE|os.O_RDWR|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	return t.Execute(f, mig)
}
