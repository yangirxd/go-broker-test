package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/digineat/go-broker-test/internal/db"
	"gitlab.com/digineat/go-broker-test/internal/services"
)

func main() {
	// Command line flags
	dbPath := flag.String("db", "data.db", "path to SQLite database")
	listenAddr := flag.String("listen", "8080", "HTTP server listen address")
	flag.Parse()

	// Initialize database connection with concurrent access parameters
	dbConn, err := sql.Open("sqlite3", *dbPath+"?_journal=WAL&_timeout=5000&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	db.InitDB(dbConn)

	repository := services.NewSqliteRepository(dbConn)

	mux := http.NewServeMux()

	mux.HandleFunc("/trades", repository.PostServerTrades())
	mux.HandleFunc("/healthz", repository.GetServerHealthz())
	mux.HandleFunc("/stats/{acc}", repository.GetServerStats())

	// Start server
	log.Printf("Starting server on %s", *listenAddr)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", *listenAddr), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
