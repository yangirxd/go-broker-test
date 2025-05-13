package main

import (
	"database/sql"
	"flag"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/digineat/go-broker-test/internal/db"
	"gitlab.com/digineat/go-broker-test/internal/services"
)

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

	tradeService := services.NewTradeService(dbConn)

	// Main worker loop
	for {
		if err := tradeService.ProcessTrades(); err != nil {
			log.Printf("Ошибка при обработке записей: %v", err)
		}
		time.Sleep(*pollInterval)
	}
}
