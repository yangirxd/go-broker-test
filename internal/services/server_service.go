package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/digineat/go-broker-test/internal/model"
)

type SqliteRepository struct {
	db *sql.DB
}

func NewSqliteRepository(db *sql.DB) *SqliteRepository {
	return &SqliteRepository{db: db}
}

// POST /trades endpoint
func (s *SqliteRepository) PostServerTrades() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		_, err := s.db.Exec(
			"INSERT INTO trades_q (account, symbol, volume, open, close, side) VALUES (?, ?, ?, ?, ?, ?)",
			trade.Account, trade.Symbol, trade.Volume, trade.Open, trade.Close, trade.Side,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to enqueue trade: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GET /healthz endpoint
func (s *SqliteRepository) GetServerHealthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := s.db.Ping(); err != nil {
			http.Error(w, "Database connection failed", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// GET /stats/{acc} endpoint
func (s *SqliteRepository) GetServerStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		account := r.URL.Path[len("/stats/"):]
		if account == "" {
			http.Error(w, "Account is required", http.StatusBadRequest)
			return
		}

		row := s.db.QueryRow("SELECT account, trades, profit FROM account_stats WHERE account = ?", account)
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
	}
}
