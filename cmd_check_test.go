package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckMigration(t *testing.T) {
	tests := map[string]struct {
		dir  string
		num  int
		diff string
	}{
		"success-create": {
			"testdata/check/success",
			10,
			"",
		},
		"success-alter-add": {
			"testdata/check/success",
			20,
			"",
		},
		"success-insert": {
			"testdata/check/success",
			30,
			"",
		},
		"success-alter-drop": {
			"testdata/check/success",
			40,
			"",
		},
		"success-nothing": {
			"testdata/check/success",
			50,
			"",
		},
		"diff-table": {
			"testdata/check/table",
			10,
			"unexpected \"table1\" table found",
		},
		"diff-schema": {
			"testdata/check/schema",
			20,
			"" +
				"create table \"table1\" differs:\n" +
				"...\n" +
				"   `val` text NOT NULL DEFAULT '',\n" +
				"+  `val2` int DEFAULT '0',\n" +
				"   PRIMARY KEY (`id`)\n" +
				"...",
		},
		"diff-record": {
			"testdata/check/record",
			30,
			"" +
				"records in \"table1\" differs:\n" +
				"+(2, 'bbb', 20)",
		},
		"diff-snapshot": {
			"testdata/check/snapshot",
			30,
			"" +
				"records in \"table1\" differs:\n" +
				" (1, 'aaa', 10)\n" +
				"-(2, 'bbb', 200)\n" +
				"+(2, 'bbb', 20)\n" +
				" (3, 'ccc', 30)",
		},
		"diff-column": {
			"testdata/check/column",
			40,
			"" +
				"records in \"table1\" differs:\n" +
				"-(1, 'aaa', 10)\n" +
				"-(2, 'bbb', 20)\n" +
				"-(3, 'ccc', 30)\n" +
				"+(1, '', 10)\n" +
				"+(2, '', 20)\n" +
				"+(3, '', 30)",
		},
		"diff-drop": {
			"testdata/check/drop",
			50,
			"missing \"table1\" table",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			diff, err := checkMigration(test.dir, test.num)
			if err != nil {
				t.Error(err)
			}
			if d := cmp.Diff(test.diff, diff); d != "" {
				t.Error(d)
			}
		})
	}
}
