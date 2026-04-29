package tests

import (
	"database/sql"
	"testing"

	"github.com/claude/blog/internal/model"
)

// openTestDB opens an in-memory SQLite database with migrations applied.
// migrationsDir is relative to the project root (TestMain already chdir's there).
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := model.Open(":memory:", "migrations")
	if err != nil {
		t.Fatalf("model.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDB_MigrationsApplied(t *testing.T) {
	db := openTestDB(t)

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least one migration to be recorded")
	}
}

func TestDB_Idempotent(t *testing.T) {
	// Opening the same in-memory DB twice is not meaningful, so we use a temp file.
	tmp := t.TempDir() + "/test.db"

	db1, err := model.Open(tmp, "migrations")
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	db1.Close()

	db2, err := model.Open(tmp, "migrations")
	if err != nil {
		t.Fatalf("second Open (idempotent): %v", err)
	}
	db2.Close()
}

func TestDB_TablesExist(t *testing.T) {
	db := openTestDB(t)

	tables := []string{
		"categories", "users", "posts", "comments", "tags",
		"post_tags", "follows", "notifications", "email_queue",
		"verification_tokens", "about_pages", "sessions", "schema_migrations",
	}

	for _, table := range tables {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		if err == sql.ErrNoRows {
			t.Errorf("table %q does not exist", table)
		} else if err != nil {
			t.Errorf("query for table %q: %v", table, err)
		}
	}
}

func TestDB_Pragmas(t *testing.T) {
	db := openTestDB(t)

	var fkEnabled int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fkEnabled); err != nil {
		t.Fatalf("query foreign_keys pragma: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("expected foreign_keys=1, got %d", fkEnabled)
	}
}
