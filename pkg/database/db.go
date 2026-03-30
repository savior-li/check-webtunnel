package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	path string
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create dir failed: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}

	return &DB{DB: db, path: dbPath}, nil
}

func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS bridges (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL UNIQUE,
		transport TEXT NOT NULL DEFAULT 'webtunnel',
		address TEXT NOT NULL,
		port INTEGER NOT NULL,
		fingerprint TEXT,
		discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_validated DATETIME,
		is_available INTEGER DEFAULT -1,
		response_time_ms INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_bridges_hash ON bridges(hash);
	CREATE INDEX IF NOT EXISTS idx_bridges_available ON bridges(is_available);
	CREATE INDEX IF NOT EXISTS idx_bridges_discovered ON bridges(discovered_at);

	CREATE TABLE IF NOT EXISTS validation_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bridge_id INTEGER NOT NULL,
		validated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_available INTEGER NOT NULL,
		response_time_ms INTEGER,
		error_message TEXT,
		FOREIGN KEY (bridge_id) REFERENCES bridges(id)
	);

	CREATE INDEX IF NOT EXISTS idx_history_bridge ON validation_history(bridge_id);
	CREATE INDEX IF NOT EXISTS idx_history_validated ON validation_history(validated_at);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Path() string {
	return db.path
}
