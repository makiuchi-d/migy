package dbstate_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/testdb"

	"github.com/makiuchi-d/migy/dbstate"
	"github.com/makiuchi-d/migy/sqlfile"
)

var testSQL = []byte(`-- test SQL
CREATE TABLE _migrations (
   id      INTEGER NOT NULL,
   applied DATETIME,
   title   VARCHAR(255),
   PRIMARY KEY (id)
);

CREATE TABLE users (
  id INTEGER NOT NULL,
  name VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
);
DELIMITER //

CREATE PROCEDURE _migration_exists(IN input_id INTEGER)
BEGIN
    IF NOT EXISTS (SELECT 1 FROM _migrations WHERE id = input_id) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'migration not found';
    END IF;
END //

\d;

INSERT INTO _migrations (id, applied, title) VALUES (1, '2025-04-19 00:33:32', 'first');
INSERT INTO users (id, name) VALUES (1, 'alice'), (2, 'bob'), (3, 'carol');
`)

func prepareTestDb(t *testing.T) *sqlx.DB {
	db := sqlx.NewDb(testdb.New("db"), "mysql")
	for s := range sqlfile.Parse(testSQL) {
		if _, err := db.Exec(s); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestGetTables(t *testing.T) {
	db := prepareTestDb(t)

	tbls, err := dbstate.GetTables(db)
	if err != nil {
		t.Fatal(err)
	}

	exp := []*dbstate.Table{
		{
			Name: "_migrations",
			Create: "" +
				"CREATE TABLE `_migrations` (\n" +
				"  `id` int NOT NULL,\n" +
				"  `applied` datetime,\n" +
				"  `title` varchar(255),\n" +
				"  PRIMARY KEY (`id`)\n" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin",
		},
		{
			Name: "users",
			Create: "" +
				"CREATE TABLE `users` (\n" +
				"  `id` int NOT NULL,\n" +
				"  `name` varchar(255) NOT NULL,\n" +
				"  PRIMARY KEY (`id`)\n" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin",
		},
	}

	if diff := cmp.Diff(tbls, exp); diff != "" {
		t.Fatal(diff)
	}
}

func TestGetProcedures(t *testing.T) {
	db := prepareTestDb(t)

	procs, err := dbstate.GetProcedures(db)
	if err != nil {
		t.Fatal(err)
	}

	exp := []*dbstate.Procedure{
		{
			Name: "_migration_exists",
			Create: "" +
				"CREATE PROCEDURE _migration_exists(IN input_id INTEGER)\n" +
				"BEGIN\n" +
				"    IF NOT EXISTS (SELECT 1 FROM _migrations WHERE id = input_id) THEN\n" +
				"        SIGNAL SQLSTATE '45000'\n" +
				"            SET MESSAGE_TEXT = 'migration not found';\n" +
				"    END IF;\n" +
				"END",
			Charset:     "utf8mb4",
			Collation:   "utf8mb4_0900_bin",
			DBCollation: "utf8mb4_0900_bin",
		},
	}

	if diff := cmp.Diff(procs, exp); diff != "" {
		t.Fatal(diff)
	}
}
