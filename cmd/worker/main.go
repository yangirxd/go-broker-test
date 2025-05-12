package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/digineat/go-broker-test/internal/db"
)

func processTrades(dbConn *sql.DB) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, account, symbol, volume, open, close, side FROM trades_q WHERE processed = 0")
	if err != nil {
		return fmt.Errorf("failed to query trades: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id      int
			account string
			symbol  string
			volume  float64
			open    float64
			close   float64
			side    string
		)

		if err := rows.Scan(&id, &account, &symbol, &volume, &open, &close, &side); err != nil {
			log.Printf("Ошибка при сканировании записи: %v", err)
			continue
		}

		multiplier := 100.0
		if strings.HasPrefix(symbol, "JPY") || strings.HasSuffix(symbol, "JPY") {
			multiplier = 100.0
		} else {
			multiplier = 10000.0
		}
		points := close - open
		if side == "sell" {
			points = -points
		}
		profit := points * volume * multiplier
		profit = float64(int64((profit)*100+0.5)) / 100

		// Обновление статистики аккаунта
		_, err = tx.Exec(
			"INSERT INTO account_stats (account, trades, profit) VALUES (?, 1, ?) "+
				"ON CONFLICT(account) DO UPDATE SET trades = trades + 1, profit = profit + excluded.profit",
			account, profit,
		)
		if err != nil {
			log.Printf("Ошибка при обновлении статистики аккаунта: %v", err)
			continue
		}

		// Пометка записи как обработанной
		_, err = tx.Exec("UPDATE trades_q SET processed = 1 WHERE id = ?", id)
		if err != nil {
			log.Printf("Ошибка при обновлении статуса записи: %v", err)
			continue
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating trades: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func main() {
	// Command line flags
	dbPath := flag.String("db", "data.db", "path to SQLite database")
	pollInterval := flag.Duration("poll", 100*time.Millisecond, "polling interval")
	flag.Parse()

	// Initialize database connection with concurrent access parameters
	dbConn, err := sql.Open("sqlite3", *dbPath+"?_journal=WAL&_timeout=5000&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	// Test database connection
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize database schema
	db.InitDB(dbConn)

	log.Printf("Worker started with polling interval: %v", *pollInterval)

	// Main worker loop
	for {
		if err := processTrades(dbConn); err != nil {
			log.Printf("Ошибка при обработке записей: %v", err)
		}
		time.Sleep(*pollInterval)
	}
}
