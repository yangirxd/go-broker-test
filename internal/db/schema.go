package db

import (
	"database/sql"
	"log"
)

const createTradesQTable = `
CREATE TABLE IF NOT EXISTS trades_q (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	account TEXT NOT NULL,
	symbol TEXT NOT NULL,
	volume REAL NOT NULL,
	open REAL NOT NULL,
	close REAL NOT NULL,
	side TEXT NOT NULL,
	processed INTEGER DEFAULT 0
);
`

const createAccountStatsTable = `
CREATE TABLE IF NOT EXISTS account_stats (
	account TEXT PRIMARY KEY,
	trades INTEGER DEFAULT 0,
	profit REAL DEFAULT 0.0
);
`

func InitDB(db *sql.DB) {
	// Включаем WAL режим и устанавливаем параметры для конкурентного доступа
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			log.Printf("Warning: Failed to set %s: %v", pragma, err)
		}
	}

	if _, err := db.Exec(createTradesQTable); err != nil {
		log.Fatalf("Failed to create trades_q table: %v", err)
	}
	log.Println("Table trades_q created or already exists")

	if _, err := db.Exec(createAccountStatsTable); err != nil {
		log.Fatalf("Failed to create account_stats table: %v", err)
	}
	log.Println("Table account_stats created or already exists")
}
