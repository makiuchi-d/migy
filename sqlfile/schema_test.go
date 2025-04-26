package sqlfile

import (
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

func TestGetTables(t *testing.T) {
	db := prepareTestDb(t)

	tbls, err := getTables(db)
	if err != nil {
		t.Fatal(err)
	}

	exp := []Table{
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

	procs, err := getProcedures(db)
	if err != nil {
		t.Fatal(err)
	}

	exp := []Procedure{
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
