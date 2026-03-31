package tests

import (
	"os"
	"path/filepath"
	"testing"

	"tor-bridge-collector/pkg/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("create test db failed: %v", err)
	}

	if err := db.InitSchema(); err != nil {
		db.Close()
		t.Fatalf("init test schema failed: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tempDir)
	}

	return db, cleanup
}

func createTestFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write test file %s failed: %v", name, err)
	}
	return path
}
