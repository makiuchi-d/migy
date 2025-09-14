package dbstate_test

import (
	"errors"
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

CREATE TABLE user_emails (
  user_id INTEGER NOT NULL,
  email   VARCHAR(255) NOT NULL,
  CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id),
  PRIMARY KEY (user_id, email)
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
INSERT INTO user_emails (user_id, email) VALUES (1, 'alice@example.com'), (2, 'bob1@example.com'), (2, 'bob2@example.com');
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

func TestHasMigrationtable(t *testing.T) {
	db := prepareTestDb(t)

	err := dbstate.HasMigrationTable(db)
	if err != nil {
		t.Fatalf("HasMigrationTable(ok): %v", err)
	}

	_, err = db.Exec("DROP TABLE _migrations")
	if err != nil {
		t.Fatalf("Drop table: %v", err)
	}

	err = dbstate.HasMigrationTable(db)
	if !errors.Is(err, dbstate.ErrNoMigrationTable) {
		t.Fatalf("HasMigrationTable(ng): %v", err)
	}
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
		{
			Name: "user_emails",
			Create: "" +
				"CREATE TABLE `user_emails` (\n" +
				"  `user_id` int NOT NULL,\n" +
				"  `email` varchar(255) NOT NULL,\n" +
				"  PRIMARY KEY (`user_id`,`email`),\n" +
				"  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)\n" +
				") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin",
			Refs: []string{"users"},
		},
	}

	if diff := cmp.Diff(exp, tbls); diff != "" {
		t.Fatal(diff)
	}
}

func TestGetTableWithForeignKeys(t *testing.T) {
	// alpha --- beta
	//        /       \
	// gamma --- delta --- epsilon
	queries := []string{
		"CREATE TABLE alpha (id int NOT NULL PRIMARY KEY)",
		"CREATE TABLE gamma (id int NOT NULL PRIMARY KEY)",
		"CREATE TABLE beta (id int NOT NULL PRIMARY KEY, aid int, gid int," +
			"FOREIGN KEY (aid) REFERENCES alpha (id)," +
			"FOREIGN KEY (gid) REFERENCES gamma (id))",
		"CREATE TABLE delta (id int NOT NULL PRIMARY KEY, gid int," +
			"FOREIGN KEY (gid) REFERENCES gamma (id))",
		"CREATE TABLE epsilon (id int NOT NULL PRIMARY KEY, bid int, did int," +
			"FOREIGN KEY (bid) REFERENCES beta (id)," +
			"FOREIGN KEY (did) REFERENCES delta (id))",
	}
	orderexp := []string{"alpha", "gamma", "beta", "delta", "epsilon"}

	db := sqlx.NewDb(testdb.New("db"), "mysql")
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}

	tbls, err := dbstate.GetTables(db)
	if err != nil {
		t.Fatal(err)
	}
	var order []string
	for _, t := range tbls {
		order = append(order, t.Name)
	}

	if diff := cmp.Diff(orderexp, order); diff != "" {
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
