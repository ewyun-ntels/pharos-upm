package store

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSQLiteStore(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite3: %v", err)
	}
	defer db.Close()

	store, err := NewSQLiteStoreWithDB(db, "logs", 10)
	if err != nil {
		t.Fatalf("failed to create SQLite store: %v", err)
	}

	runStoreTests(t, store)
}
