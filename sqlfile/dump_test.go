package sqlfile

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
