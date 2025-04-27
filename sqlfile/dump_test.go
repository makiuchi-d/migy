package sqlfile

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"
)

func prepareTestDb(t *testing.T) *sqlx.DB {
	db := sqlx.NewDb(testdb.New("db"), "mysql")
	for s := range Parse(testSQL) {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestDump(t *testing.T) {
	exp, err := os.ReadFile("testdata/dump/golden.sql")
	if err != nil {
		t.Fatal(err)
	}

	db := prepareTestDb(t)
	buf := bytes.NewBuffer(nil)

	err = Dump(buf, db)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(buf.String(), string(exp)); diff != "" {
		t.Fatal(diff)
	}
}
