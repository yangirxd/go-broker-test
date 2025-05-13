package services

import (
	"database/sql"
	"testing"
)

func SetupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось открыть базу данных в памяти: %v", err)
	}

	createTradesQ := `
		CREATE TABLE trades_q (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account TEXT NOT NULL,
			symbol TEXT NOT NULL,
			volume REAL NOT NULL,
			open REAL NOT NULL,
			close REAL NOT NULL,
			side TEXT NOT NULL,
			processed INTEGER DEFAULT 0
		)`
	createAccountStats := `
		CREATE TABLE account_stats (
			account TEXT PRIMARY KEY,
			trades INTEGER DEFAULT 0,
			profit REAL DEFAULT 0.0
		)`

	if _, err := db.Exec(createTradesQ); err != nil {
		t.Fatalf("Не удалось создать таблицу trades_q: %v", err)
	}
	if _, err := db.Exec(createAccountStats); err != nil {
		t.Fatalf("Не удалось создать таблицу account_stats: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
