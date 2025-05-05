package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Command line flags
	dbPath := flag.String("db", "data.db", "path to SQLite database")
	listenAddr := flag.String("listen", "8080", "HTTP server listen address")
	flag.Parse()

	// Initialize database connection
	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Initialize HTTP server
	mux := http.NewServeMux()

	// POST /trades endpoint
	mux.HandleFunc("POST /trades", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Write code here
		w.WriteHeader(http.StatusOK)
	})

	// GET /stats/{acc} endpoint
	mux.HandleFunc("GET /stats/{acc}", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Write code here
		w.WriteHeader(http.StatusOK)
	})

	// GET /healthz endpoint
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Write code here
		// 1. Check database connection
		// 2. Return health status
		w.WriteHeader(http.StatusOK)
	})

	// Start server
	serverAddr := fmt.Sprintf(":%s", *listenAddr)
	log.Printf("Starting server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
