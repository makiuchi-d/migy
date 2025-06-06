package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// cmdInit represents the init command
var cmdInit = &cobra.Command{
	Use:   "init",
	Short: "Generate the initial migration SQL file",
	Long: `Generate the initial migration SQL file.
This file sets up the initial state of the database,
including the _migrations table used to track applied migrations.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		return generateInitSQLFile(targetDir, overwrite)
	},
}

func init() {
	cmd.AddCommand(cmdInit)
	addFlagForce(cmdInit)
}

const initFile = "000000_init.all.sql"
const initSQL = signature + `

CREATE TABLE _migrations (
   id      INTEGER NOT NULL,
   applied DATETIME,
   title   VARCHAR(255),
   PRIMARY KEY (id)
);

INSERT INTO _migrations (id, applied, title) VALUES (0, now(), 'init');

DELIMITER //

CREATE PROCEDURE _migration_exists(IN input_id INTEGER)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM _migrations WHERE id = input_id) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'migration not found';
  END IF;
END//

DELIMITER ;
`

func generateInitSQLFile(dir string, overwrite bool) error {
	path := filepath.Join(dir, initFile)
	flag := os.O_CREATE | os.O_RDWR | os.O_TRUNC
	if !overwrite {
		flag |= os.O_EXCL
	}
	f, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(initSQL))
	return err
}
