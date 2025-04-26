package sqlfile

import (
	"os"

	"github.com/jmoiron/sqlx"
)

// Apply applies SQL file to db.DB
func Apply(db sqlx.Execer, file string) error {
	input, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	for s := range Parse(input) {
		_, err := db.Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}
