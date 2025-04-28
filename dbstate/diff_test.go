package dbstate

import (
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/anydiff"
	"github.com/makiuchi-d/testdb"
)

func TestDiffTables(t *testing.T) {
	var (
		setupSQLs = []string{`
CREATE PROCEDURE _migration_exists()
BEGIN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'dummy';
END
`, `
CREATE TABLE users (
    id INTEGER NOT NULL,
    name VARCHAR(255),
    age  INTEGER,
    PRIMARY KEY (id))
`, `
INSERT INTO users (id, name, age) VALUES
    (1, 'alice', 30), (2, 'bob', 24), (3, 'carol', null)
`, `
CREATE TABLE table1 (
    id   INTEGER NOT NULL,
    val  INTEGER NOT NULL,
    PRIMARY KEY (id)
)`, `
CREATE TABLE table2 (
    id  INTEGER NOT NULL,
    val INTEGER NOT NULL,
    PRIMARY KEY (id)
)`,
		}
		changeSQLs = []string{
			`UPDATE users SET age = 28 WHERE id = 3`,
			`DELETE FROM users WHERE id = 2`,
			`ALTER TABLE table1 ADD COLUMN val2 TEXT AFTER val`,
			`DROP TABLE table2`,
			`
CREATE TABLE table3 (
    id  INTEGER NOT NULL,
    val TEXT,
    PRIMARY KEY (id)
)`,
		}
	)

	db := sqlx.NewDb(testdb.New("db"), "myseql")

	for _, sql := range setupSQLs {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}

	ss, err := TakeSnapshot(db)
	if err != nil {
		t.Fatal(err)
	}

	for _, sql := range changeSQLs {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}

	var sb strings.Builder
	err = diffTables(&sb, db, ss)
	if err != nil {
		t.Fatal(err)
	}

	exp := "" +
		"create table \"table1\" differs:\n" +
		"...\n" +
		"   `val` int NOT NULL,\n" +
		"+  `val2` text,\n" +
		"   PRIMARY KEY (`id`)\n" +
		"...\n" +
		"unexpected \"table3\" table found\n" +
		"missing \"table2\" table\n"

	diff := sb.String()
	if diff != exp {
		t.Errorf("diff:\n%v\nwants:\n%v", diff, exp)
	}
}

func TestDiffProcedures(t *testing.T) {
	// When a snapshot is taken, only ‘_migration_exists’ can be retrieved.
	// see GetProcedures()
	var (
		setupSQLs = []string{`
CREATE PROCEDURE _migration_exists()
BEGIN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'proc0';
END
`,
		}
		changeSQLs = []string{
			`DROP PROCEDURE _migration_exists`,
			`
CREATE PROCEDURE _migration_exists()
BEGIN
    SIGNAL SQLSTATE '45001' SET MESSAGE_TEXT = 'proc1';
END`,
		}
	)

	db := sqlx.NewDb(testdb.New("db"), "mysql")

	for _, sql := range setupSQLs {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}

	ss, err := TakeSnapshot(db)
	if err != nil {
		t.Fatal(err)
	}

	for _, sql := range changeSQLs {
		if _, err := db.Exec(sql); err != nil {
			t.Fatal(err)
		}
	}

	var sb strings.Builder
	err = diffProcedures(&sb, db, ss)
	if err != nil {
		t.Fatal(err)
	}

	exp := `stored procedure "_migration_exists" differs:
...
 BEGIN
-    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'proc0';
+    SIGNAL SQLSTATE '45001' SET MESSAGE_TEXT = 'proc1';
 END
`
	diff := sb.String()
	if diff != exp {
		t.Errorf("diff:\n%v\nwants:\n%v", diff, exp)
	}
}

func TestDiffString(t *testing.T) {
	tests := map[string]struct {
		a   []int
		b   []int
		exp string
	}{
		"1st": {
			[]int{2, 3, 4, 5},
			[]int{1, 2, 3, 4, 5},
			"+1\n 2\n...\n",
		},
		"2nd": {
			[]int{1, 2, 3, 4, 5},
			[]int{1, 3, 4, 5},
			" 1\n-2\n 3\n...\n",
		},
		"center": {
			[]int{1, 2, 3, 4, 5},
			[]int{1, 2, 4, 5},
			"...\n 2\n-3\n 4\n...\n",
		},
		"penult": {
			[]int{1, 2, 3, 5},
			[]int{1, 2, 3, 4, 5},
			"...\n 3\n+4\n 5\n",
		},
		"last": {
			[]int{1, 2, 3, 4, 5},
			[]int{1, 2, 3, 4},
			"...\n 4\n-5\n",
		},
		"skip1": {
			[]int{1, 2, 3, 5},
			[]int{1, 3, 4, 5},
			" 1\n-2\n 3\n+4\n 5\n",
		},
		"skip2": {
			[]int{1, 2, 3, 4},
			[]int{1, 3, 4, 5},
			" 1\n-2\n 3\n 4\n+5\n",
		},
		"split": {
			[]int{1, 2, 3, 4},
			[]int{2, 3, 4, 5},
			"-1\n 2\n...\n 4\n+5\n",
		},
	}
	for k, test := range tests {
		t.Run(k, func(t *testing.T) {
			edit := anydiff.Diff(test.a, test.b, anydiff.Cmp)
			var sb strings.Builder
			diffString(&sb, edit, test.a, test.b)
			diff := sb.String()
			if diff != test.exp {
				t.Fatalf("diff:\n%v\nwants:\n%v", diff, test.exp)
			}
		})
	}

	a := []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
		"line10",
	}
	b := []string{
		"line1",
		"line3",
		"line4",
		"line5",
		"line6.5",
		"line7",
		"line8",
		"line8.5",
		"line9",
		"line10",
	}

	exp := ` line1
-line2
 line3
...
 line5
-line6
+line6.5
 line7
 line8
+line8.5
 line9
...
`

	edit := anydiff.Diff(a, b, anydiff.Cmp)

	var sb strings.Builder
	diffString(&sb, edit, a, b)
	diff := sb.String()

	if diff != exp {
		t.Fatalf("diff:\n%v\nwants\n%v", diff, exp)
	}
}
