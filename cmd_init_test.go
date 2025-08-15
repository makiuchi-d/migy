package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateInitSQLFile(t *testing.T) {
	tests := map[string]struct {
		exist     bool
		overwrite bool
		wantErr   bool
	}{
		"success": {
			exist:     false,
			overwrite: false,
			wantErr:   false,
		},
		"already exists": {
			exist:     true,
			overwrite: false,
			wantErr:   true,
		},
		"overwrite": {
			exist:     true,
			overwrite: true,
			wantErr:   false,
		},
	}

	golden, err := os.ReadFile(filepath.Join("testdata", "init", "golden.sql"))
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "000000_init.all.sql")

			if tt.exist {
				if err := os.WriteFile(path, []byte("dummy content"), 0644); err != nil {
					t.Fatalf("failed to create dummy file: %v", err)
				}
			}

			err := generateInitSQLFile(tmpDir, tt.overwrite)

			if (err != nil) != tt.wantErr {
				t.Fatalf("generateInitSQLFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("could not read init file: %v", err)
			}
			if string(content) != string(golden) {
				t.Errorf("init file content is wrong: got %s, want %s", content, golden)
			}
		})
	}
}
