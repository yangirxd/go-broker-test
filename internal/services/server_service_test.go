package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/digineat/go-broker-test/internal/model"
)

func TestPostServerTrades(t *testing.T) {
	dbConn, cleanup := SetupTestDB(t)
	defer cleanup()
	repo := NewSqliteRepository(dbConn)
	handler := repo.PostServerTrades()

	t.Run("valid trade", func(t *testing.T) {
		trade := model.Trade{Account: "ACC1", Symbol: "EURUSD", Volume: 1.5, Open: 1.2345, Close: 1.2350, Side: "buy"}
		body, _ := json.Marshal(trade)
		req := httptest.NewRequest(http.MethodPost, "/trades", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusNoContent {
			t.Errorf("Expected 204, got %d", rr.Code)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/trades", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405, got %d", rr.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/trades", bytes.NewReader([]byte("invalid")))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
	})

	t.Run("invalid trade", func(t *testing.T) {
		trade := model.Trade{Account: "", Symbol: "EURUSD", Volume: 1.5, Open: 1.2345, Close: 1.2350, Side: "buy"}
		body, _ := json.Marshal(trade)
		req := httptest.NewRequest(http.MethodPost, "/trades", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
	})
}

func TestGetServerHealthz(t *testing.T) {
	dbConn, cleanup := SetupTestDB(t)
	defer cleanup()
	repo := NewSqliteRepository(dbConn)
	handler := repo.GetServerHealthz()

	t.Run("health ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
		if rr.Body.String() != "OK" {
			t.Errorf("Expected body OK, got %s", rr.Body.String())
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405, got %d", rr.Code)
		}
	})

	t.Run("db closed", func(t *testing.T) {
		dbConn.Close()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %d", rr.Code)
		}
	})
}

func TestGetServerStats(t *testing.T) {
	dbConn, cleanup := SetupTestDB(t)
	defer cleanup()
	repo := NewSqliteRepository(dbConn)
	handler := repo.GetServerStats()

	t.Run("existing account", func(t *testing.T) {
		_, err := dbConn.Exec("INSERT INTO account_stats (account, trades, profit) VALUES ('ACC1', 5, 1000.5)")
		if err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}
		req := httptest.NewRequest(http.MethodGet, "/stats/ACC1", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp["account"] != "ACC1" || resp["trades"] != float64(5) || resp["profit"] != 1000.5 {
			t.Errorf("Unexpected response: %v", resp)
		}
	})

	t.Run("non-existing account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/ACC2", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", rr.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(rr.Body).Decode(&resp)
		if resp["account"] != "ACC2" || resp["trades"] != float64(0) || resp["profit"] != 0.0 {
			t.Errorf("Unexpected response: %v", resp)
		}
	})

	t.Run("empty account", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/stats/", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", rr.Code)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stats/ACC1", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected 405, got %d", rr.Code)
		}
	})

	t.Run("db closed", func(t *testing.T) {
		dbConn.Close()
		req := httptest.NewRequest(http.MethodGet, "/stats/ACC1", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected 500, got %d", rr.Code)
		}
	})
}
