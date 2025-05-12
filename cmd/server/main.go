package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/digineat/go-broker-test/internal/db"
	"gitlab.com/digineat/go-broker-test/internal/model"
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

	mux := http.NewServeMux()

	// POST /trades endpoint
	mux.HandleFunc("/trades", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var trade model.Trade
		if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if err := model.ValidateTrade(trade); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := dbConn.Exec(
			"INSERT INTO trades_q (account, symbol, volume, open, close, side) VALUES (?, ?, ?, ?, ?, ?)",
			trade.Account, trade.Symbol, trade.Volume, trade.Open, trade.Close, trade.Side,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to enqueue trade: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	// GET /healthz endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := dbConn.Ping(); err != nil {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// GET /stats/{acc} endpoint
	mux.HandleFunc("/stats/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		account := r.URL.Path[len("/stats/"):]
		if account == "" {
			http.Error(w, "Account is required", http.StatusBadRequest)
			return
		}

		row := dbConn.QueryRow("SELECT account, trades, profit FROM account_stats WHERE account = ?", account)
		var (
			acc    string
			trades int
			profit float64
		)
		if err := row.Scan(&acc, &trades, &profit); err != nil {
			if err == sql.ErrNoRows {
				// Если аккаунт не найден, возвращаем нулевые значения
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"account": account,
					"trades":  0,
					"profit":  0.0,
				})
				return
			}
			http.Error(w, fmt.Sprintf("Failed to fetch stats: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"account": acc,
			"trades":  trades,
			"profit":  profit,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Start server
	log.Printf("Starting server on %s", *listenAddr)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", *listenAddr), mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
