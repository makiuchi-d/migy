package main

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/migy/sqlfile"
	"github.com/makiuchi-d/testdb"
)

func TestListFilesToApply(t *testing.T) {
	var dir = filepath.Join("testdata", "list")

	dbempty := sqlx.NewDb(testdb.New("db"), "mysql")
	db30 := sqlx.NewDb(testdb.New("db"), "mysql")
	if err := sqlfile.Apply(db30, filepath.Join(dir, "000030_third.all.sql")); err != nil {
		t.Fatalf("apply 30: %v", err)
	}

	tests := map[string]struct {
		db  *sqlx.DB
		num int
		exp []string
	}{
		"empty->20": {
			dbempty,
			20,
			[]string{
				"000000_init.all.sql",
				"000010_first.up.sql",
				"000020_second.up.sql",
			},
		},
		"empty->30": {
			dbempty,
			30,
			[]string{
				"000030_third.all.sql",
			},
		},
		"empty->50": {
			dbempty,
			50,
			[]string{
				"000030_third.all.sql",
				"000040_fourth.up.sql",
				"000050_fifth.up.sql",
			},
		},
		"30->50": {
			db30,
			50,
			[]string{
				"000040_fourth.up.sql",
				"000050_fifth.up.sql",
			},
		},
		"30->10": {
			db30,
			10,
			[]string{
				"000030_third.down.sql",
				"000020_second.down.sql",
			},
		},
		"30->30": {
			db30,
			30,
			nil,
		},
	}
	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			files, err := listFilesToApply(test.db, dir, test.num)
			if err != nil {
				t.Errorf("error: %v", err)
			}
			if diff := cmp.Diff(test.exp, files); diff != "" {
				t.Errorf("differs:\n%v", diff)
			}
		})
	}
}
