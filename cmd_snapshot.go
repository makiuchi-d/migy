package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"
	"github.com/spf13/cobra"
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

	migs, err := GetMigrations(dir)
	if err != nil {
		return err
	}

	if num == 0 {
		num = migs[len(migs)-1].Number
	}

	migs = migs.FromSnapshotTo(num)
	if len(migs) == 0 {
		return fmt.Errorf("no migrations (number: %06d)", num)
	}

	db := testdb.New("db")

	fmt.Println("applying...")

	for i, m := range migs {
		name := m.UpName()
		if i == 0 && m.Snapshot {
			name = m.SnapshotName()
		}

		fmt.Println(name)
		err := applySQLFile(db, filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	last := migs[len(migs)-1]
	last.Snapshot = true
	fmt.Println("======\nwriting:", last.SnapshotName())
	flag := os.O_CREATE | os.O_RDWR
	if !overwrite {
		flag |= os.O_EXCL
	}
	f, err := os.OpenFile(filepath.Join(dir, last.SnapshotName()), flag, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	return Dump(f, sqlx.NewDb(db, "mysql"))
}
