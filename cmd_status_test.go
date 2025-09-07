package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"

	"github.com/makiuchi-d/migy/migrations"
	"github.com/makiuchi-d/migy/sqlfile"
)

func TestReadStatus(t *testing.T) {
	dir := filepath.Join("testdata", "status")

	tests := map[string]struct {
		sqls []string
		exp  string
	}{
		"nil": {
			sqls: nil,
			exp: `
000000	　　⏺	　0000-00-00 --:--:--	"init"
000010	⏫⏬　	　0000-00-00 --:--:--	"first"
000030	⏫⏬⏺	　0000-00-00 --:--:--	"third"
`[1:],
		},
		"empty": {
			sqls: []string{},
			exp: `
000000	　　⏺	　0000-00-00 --:--:--	"init"
000010	⏫⏬　	　0000-00-00 --:--:--	"first"
000030	⏫⏬⏺	　0000-00-00 --:--:--	"third"
`[1:],
		},
		"30": {
			sqls: []string{"000030_third.all.sql"},
			exp: `
000000	　　⏺	✅2025-09-07 19:07:50	"init"
000010	⏫⏬　	✅2025-09-07 19:07:50	⚠"first" DB:"firstdb"
000020	　　　	✅2025-09-07 19:07:50	"second"
000030	⏫⏬⏺	✅2025-09-07 19:08:47	"third"
`[1:],
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var db *sqlx.DB

			if test.sqls != nil {
				db = sqlx.NewDb(testdb.New("db"), "mysql")
				defer db.Close()

				for _, f := range test.sqls {
					err := sqlfile.Apply(db, filepath.Join(dir, f))
					if err != nil {
						t.Fatal(err)

					}
				}
			}

			st, err := readStatus(db, dir)
			if err != nil {
				t.Error(err)
				return
			}
			if diff := cmp.Diff(test.exp, st); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestFormatStatus(t *testing.T) {
	newMig := func(num int, title string, updown, snapshot bool) *migrations.Migration {
		return &migrations.Migration{
			Number:   num,
			Title:    title,
			UpDown:   updown,
			Snapshot: snapshot,
		}
	}

	tests := map[string]struct {
		st  migrations.Status
		exp string
	}{
		"init": {
			st: migrations.Status{
				Migration: newMig(0, "init", false, true),
				Applied:   time.Time{},
			},
			exp: "000000\t　　⏺\t　0000-00-00 --:--:--\t\"init\"",
		},
		"applied": {
			st: migrations.Status{
				Migration: newMig(10, "first", true, false),
				Applied:   time.Date(2025, 9, 4, 10, 20, 30, 0, time.UTC),
			},
			exp: "000010\t⏫⏬　\t✅2025-09-04 10:20:30\t\"first\"",
		},
		"with-snapshot": {
			st: migrations.Status{
				Migration: newMig(20, "second", true, true),
				Applied:   time.Date(2025, 9, 5, 20, 30, 40, 0, time.UTC),
				DBTitle:   "db-second",
			},
			exp: "000020\t⏫⏬⏺\t✅2025-09-05 20:30:40\t⚠\"second\" DB:\"db-second\"",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ret := formatStatus(nil, test.st)
			if diff := cmp.Diff(test.exp, string(ret)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
