package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

// Open opens (or creates) the SQLite database at dbPath, applies WAL-mode
// pragmas, and runs any pending migrations from migrationsDir.
func Open(dbPath, migrationsDir string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("set %s: %w", pragma, err)
		}
	}

	if err := runMigrations(db, migrationsDir); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB, dir string) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename   TEXT PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE filename = ?`, name).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (filename) VALUES (?)`, name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		log.Printf("applied migration: %s", name)
	}

	return nil
}
