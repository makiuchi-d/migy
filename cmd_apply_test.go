package main

import (
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/makiuchi-d/migy/migrations"
	"github.com/makiuchi-d/testdb"
)

func TestApplyMigrations(t *testing.T) {
	targetDir := filepath.Join("testdata", "apply")

	db := sqlx.NewDb(testdb.New("db"), "mysql")

	tests := []struct {
		name    string
		num     int
		confirm bool
		ok      bool
		result  int
	}{
		{
			name:    "empty->20",
			num:     20,
			confirm: true,
			ok:      true,
			result:  20,
		},
		{
			name:    "20->30(latest)",
			num:     -1,
			confirm: true,
			ok:      true,
			result:  30,
		},
		{
			name:    "30->10",
			num:     10,
			confirm: true,
			ok:      true,
			result:  10,
		},
		{
			name:    "10->10 (nothing to do)",
			num:     10,
			confirm: true,
			ok:      true,
			result:  10,
		},
		{
			name:    "cancel (10->30)",
			num:     30,
			confirm: false,
			ok:      false,
			result:  10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ok, err := applyMigrations(db, targetDir, test.num, func(func()) bool { return test.confirm })
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if ok != test.ok {
				t.Fatalf("return %v, wants %v", ok, test.ok)
			}

			hs, err := migrations.LoadHistories(db)
			if err != nil {
				t.Fatalf("history error: %v", err)
			}
			if num := hs.CurrentNum(); num != test.result {
				t.Fatalf("current num = %v, wants %v", num, test.result)
			}
		})
	}
}
