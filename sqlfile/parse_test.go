package sqlfile

import (
	"testing"
)

var testSQL = []byte(`-- test SQL

/*
 * SCHEMA
 */

CREATE TABLE _migrations (
   id      INTEGER NOT NULL,
   applied DATETIME,
   title   VARCHAR(255),
   PRIMARY KEY (id)
);;

CREATE TABLE users (
  id INTEGER NOT NULL,
  name VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
);

/*
 * STORED PROCEDURE
 */

DELIMITER // -- make delimiter "//"

CREATE PROCEDURE _migration_exists(IN input_id INTEGER)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM _migrations WHERE id = input_id) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'migration not found';
  END IF;
END //

\d; --/* return to default delimiter ';'

/*
 * TEST DATA
 */

INSERT INTO _migrations (id, applied, title) VALUES (1, '2025-04-19 00:33:32', 'first');
INSERT INTO users (id, name) VALUES (1, 'alice'), (2, 'bob'), (3, 'carol');

-- end of file
`)

func TestSkipSpaces(t *testing.T) {
	tests := map[string]struct {
		src string
		exp int
	}{
		"empty": {
			src: "  \t  \r\n  \n\r\n\n",
			exp: len("  \t  \r\n  \n\r\n\n"),
		},
		"skipped": {
			src: "   \nselect \n   \n   ",
			exp: len("   \n"),
		},
		"unskipped": {
			src: "select \n   \n   ",
			exp: 0,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := skipSpaces([]byte(test.src))
			if n != test.exp {
				t.Errorf("%q => %v wants %v", test.src, n, test.exp)
			}
		})
	}
}

func TestSinglelineComment(t *testing.T) {
	tests := map[string]struct {
		src string
		exp int
	}{
		"abc-def":  {"-- abc\ndef\n", len("-- abc\n")},
		"newline":  {"--\n\nabc", len("--\n")},
		"lastline": {"-- abcdefg", len("-- abcdefg")},
		"empty":    {"--", 2},
		"no1":      {"/--", 0},
		"no2":      {"-/-", 0},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := singlelineComment([]byte(test.src))
			if n != test.exp {
				t.Errorf("length = %v wants %v", n, test.exp)
			}
		})
	}
}

func TestMultilineComment(t *testing.T) {
	tests := map[string]struct {
		src string
		exp int
	}{
		"abc-def":  {"/* abc\n * def\n */ghi\n", len("/* abc\n * def\n */")},
		"newline":  {"/*\n\n*/\n\n", len("/*\n\n*/")},
		"lastline": {"/* abcdefg", len("/* abcdefg")},
		"nop":      {"/**/", len("/**/")},
		"slash":    {"/*/*/*/", len("/*/*/")},
		"no1":      {"-/*", 0},
		"no2":      {"/-*", 0},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := multilineComment([]byte(test.src))
			if n != test.exp {
				t.Errorf("length = %v wants %v", n, test.exp)
			}
		})
	}
}

func TestDelimiter(t *testing.T) {
	tests := map[string]struct {
		src, delim string
		exp        bool
	}{
		"semicolon ng": {"select 1;", ";", false},
		"semicolon ok": {"; select 1;", ";", true},
		"slash ng":     {"; select 1;", "//", false},
		"slash ok":     {"// select 1;", "//", true},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			yes := delimiter([]byte(test.src), []byte(test.delim))
			if yes != test.exp {
				t.Errorf("is delimiter: %v wants %v", yes, test.exp)
			}
		})
	}
}

func TestQuotedLiteral(t *testing.T) {
	tests := map[string]struct {
		src string
		exp int
	}{
		"notquote": {"a'b\"c`def", 0},
		"single":   {"'a\"b\\'c`d'ef", 10},
		"double":   {"\"a\\\"b'c`d\"ef", 10},
		"back":     {"`a\"b'c\\`d`ef", 10},
		"empty":    {"''abc", 2},
		"unclosed": {"'abc\"def", 8},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			n := quotedLiteral([]byte(test.src))
			if n != test.exp {
				t.Errorf("length = %v wants %v", n, test.exp)
			}
		})
	}
}

func TestChangeDelimiter(t *testing.T) {
	tests := map[string]struct {
		src   string
		delim string
		len   int
	}{
		"notcmd1":  {"select", "", 0},
		"notcmd2":  {"delimiteraa", "", 0},
		"notcmd3":  {"delimiter'aa'", "", 0},
		"short1":   {"\\d//\n", "//", 4},
		"short2":   {"\\d\t XXX -- comment", "XXX", 7},
		"long":     {"Delimiter //\n", "//", 12},
		"quoted":   {"dElImItEr\t'///'select\n", "///", 15},
		"unclosed": {"delimiter 'abc", "abc", 14},
		"space":    {"delimiter ' /'  ", " /", 14},
		"newline":  {"delimiter '//\nabc'", "//", 13},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			delim, len := changeDelimiter([]byte(test.src))
			if len != test.len {
				t.Errorf("length = %v wants %v", len, test.len)
			} else if string(delim) != test.delim {
				t.Errorf("delimiter = %q wants %q", delim, test.delim)
			}
		})
	}
}

func TestParse(t *testing.T) {
	exp := []string{
		`CREATE TABLE _migrations (
   id      INTEGER NOT NULL,
   applied DATETIME,
   title   VARCHAR(255),
   PRIMARY KEY (id)
)`,
		`CREATE TABLE users (
  id INTEGER NOT NULL,
  name VARCHAR(255) NOT NULL,
  PRIMARY KEY (id)
)`,
		`CREATE PROCEDURE _migration_exists(IN input_id INTEGER)
BEGIN
  IF NOT EXISTS (SELECT 1 FROM _migrations WHERE id = input_id) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'migration not found';
  END IF;
END `,

		`INSERT INTO _migrations (id, applied, title) VALUES (1, '2025-04-19 00:33:32', 'first')`,
		`INSERT INTO users (id, name) VALUES (1, 'alice'), (2, 'bob'), (3, 'carol')`,
	}

	n := 0
	for stmt := range Parse(testSQL) {
		t.Logf("stmt[%v]\n%v\n", n, stmt)
		if n >= len(exp) {
			t.Fatalf("unexpected: %v", stmt)
			break
		}
		if stmt != exp[n] {
			t.Fatalf("statement[%v]:\n%v\n---wants:\n%v\n", n, stmt, exp[n])
		}
		n++
	}
	if n != len(exp) {
		t.Fatalf("not parsed: %v", exp[n:])
	}
}
