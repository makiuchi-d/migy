package dbstate_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/makiuchi-d/migy/dbstate"
)

func TestDiff(t *testing.T) {
	db := prepareTestDb(t)

	ss, err := dbstate.TakeSnapshot(db)
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}

	diff, err := dbstate.Diff(db, ss, nil)
	if err != nil {
		t.Errorf("Diff(1) error: %v", err)
	}
	if d := cmp.Diff(diff, ""); d != "" {
		t.Errorf("Diff(2):\n%v\n", d)
	}

	for _, q := range []string{
		"ALTER TABLE users ADD age int AFTER `name`",
		"INSERT INTO user_emails (user_id, email) VALUES (3, 'carol@example.com')",
	} {
		_, err := db.Exec(q)
		if err != nil {
			t.Fatalf("db.Exec(%v): %v", q, err)
		}
	}

	exp := "" +
		"create table \"users\" differs:\n" +
		"...\n" +
		"   `name` varchar(255) NOT NULL,\n" +
		"+  `age` int,\n" +
		"   PRIMARY KEY (`id`)\n" +
		"...\n" +
		"records in \"user_emails\" differs:\n" +
		"...\n" +
		" (2, 'bob2@example.com')\n" +
		"+(3, 'carol@example.com')\n"

	diff, err = dbstate.Diff(db, ss, nil)
	if err != nil {
		t.Errorf("Diff(2) error: %v", err)
	}
	if d := cmp.Diff(diff, exp); d != "" {
		t.Errorf("Diff(2):\n%v\n", d)
	}
}
