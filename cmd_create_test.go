package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextNum(t *testing.T) {
	tests := map[int]int{
		0:  10,
		1:  10,
		9:  10,
		10: 20,
		11: 20,
		19: 20,
		20: 30,
	}
	for n, exp := range tests {
		r := nextNum(n)
		if r != exp {
			t.Errorf("nextNum(%d) = %d wants %d", n, r, exp)
		}
	}
}

func TestCreateNewMigrationFiles(t *testing.T) {
	tests := map[string]struct {
		num      int
		title    string
		files    []string
		wantErr  bool
		wantNum  int
		wantUp   string
		wantDown string
	}{
		"initial": {
			num:      -1,
			title:    "first_migration",
			files:    []string{"000000_init.all.sql"},
			wantNum:  10,
			wantUp:   "VALUES (10, 'first_migration'",
			wantDown: "CALL _migration_exists(10)",
		},
		"subsequent": {
			num:      -1,
			title:    "second_migration",
			files:    []string{"000000_init.all.sql", "000010_first.up.sql", "000010_first.down.sql"},
			wantNum:  20,
			wantUp:   "VALUES (20, 'second_migration'",
			wantDown: "CALL _migration_exists(20)",
		},
		"specific number": {
			num:      5,
			title:    "specific_migration",
			files:    []string{"000000_init.all.sql", "000010_first.up.sql", "000010_first.down.sql"},
			wantNum:  5,
			wantUp:   "VALUES (5, 'specific_migration'",
			wantDown: "CALL _migration_exists(5)",
		},
		"duplicate": {
			num:     10,
			title:   "duplicate_migration",
			files:   []string{"000000_init.all.sql", "000010_first.up.sql", "000010_first.down.sql"},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for _, file := range tt.files {
				p := filepath.Join(tmpDir, file)
				if err := os.WriteFile(p, []byte("-- dummy"), 0644); err != nil {
					t.Fatalf("setup failed to create dummy file %s: %v", file, err)
				}
			}

			err := createNewMigrationFiles(tmpDir, tt.num, tt.title)

			if (err != nil) != tt.wantErr {
				t.Fatalf("createNewMigrationFiles() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			upFile := fmt.Sprintf("%06d_%s.up.sql", tt.wantNum, tt.title)
			downFile := fmt.Sprintf("%06d_%s.down.sql", tt.wantNum, tt.title)

			upPath := filepath.Join(tmpDir, upFile)
			if _, err := os.Stat(upPath); err != nil {
				t.Errorf("up file was not created: %v", err)
			}

			downPath := filepath.Join(tmpDir, downFile)
			if _, err := os.Stat(downPath); err != nil {
				t.Errorf("down file was not created: %v", err)
			}

			content, err := os.ReadFile(upPath)
			if err != nil {
				t.Fatalf("could not read up file: %v", err)
			}
			if !strings.Contains(string(content), tt.wantUp) {
				t.Errorf("up file content is wrong: got %s, want %s", content, tt.wantUp)
			}

			content, err = os.ReadFile(downPath)
			if err != nil {
				t.Fatalf("could not read down file: %v", err)
			}
			if !strings.Contains(string(content), tt.wantDown) {
				t.Errorf("down file content is wrong: got %s, want %s", content, tt.wantDown)
			}
		})
	}
}
