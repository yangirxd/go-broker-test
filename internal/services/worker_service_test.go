package services

import (
	"database/sql"
	"math"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
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

func TestProcessTrades_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tradeService := NewTradeService(db)

	_, err := db.Exec(`
		INSERT INTO trades_q (account, symbol, volume, open, close, side)
		VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?),
		(?, ?, ?, ?, ?, ?)`,
		"user1", "BTC", 1.5, 100.0, 110.0, "buy",
		"user1", "ETH", 2.0, 200.0, 190.0, "sell",
		"user2", "BTC", 1.0, 150.0, 160.0, "buy")
	if err != nil {
		t.Fatalf("Не удалось вставить тестовые данные: %v", err)
	}

	err = tradeService.ProcessTrades()
	if err != nil {
		t.Fatalf("Ожидалось успешное выполнение ProcessTrades, но произошла ошибка: %v", err)
	}

	var processedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM trades_q WHERE processed = 1").Scan(&processedCount)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к trades_q: %v", err)
	}
	if processedCount != 3 {
		t.Errorf("Ожидалось, что обработано 3 записи, но обработано: %d", processedCount)
	}

	var trades1 int
	var profit1 float64
	err = db.QueryRow("SELECT trades, profit FROM account_stats WHERE account = 'user1'").Scan(&trades1, &profit1)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к account_stats для user1: %v", err)
	}
	if trades1 != 2 {
		t.Errorf("Ожидалось 2 сделки для user1, но найдено: %d", trades1)
	}
	expectedProfit1 := 3500000.0
	if math.Abs(profit1-expectedProfit1) > 0.01 {
		t.Errorf("Ожидалась прибыль для user1 %.1f, но найдено: %.1f", expectedProfit1, profit1)
	}

	var trades2 int
	var profit2 float64
	err = db.QueryRow("SELECT trades, profit FROM account_stats WHERE account = 'user2'").Scan(&trades2, &profit2)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к account_stats для user2: %v", err)
	}
	if trades2 != 1 {
		t.Errorf("Ожидалась 1 сделка для user2, но найдено: %d", trades2)
	}
	expectedProfit2 := 1000000.0
	if math.Abs(profit2-expectedProfit2) > 0.01 {
		t.Errorf("Ожидалась прибыль для user2 %.1f, но найдено: %.1f", expectedProfit2, profit2)
	}
}

func TestProcessTrades_ErrorOnBegin(t *testing.T) {
	db, cleanup := setupTestDB(t)
	cleanup()

	tradeService := NewTradeService(db)

	err := tradeService.ProcessTrades()
	if err == nil {
		t.Error("Ожидалась ошибка при начале транзакции, но её не произошло")
	}
	if err.Error() != "failed to begin transaction: sql: database is closed" {
		t.Errorf("Ожидалась ошибка 'failed to begin transaction: sql: database is closed', но получено: %v", err)
	}
}

func TestProcessTrades_RoundProfit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tradeService := NewTradeService(db)

	_, err := db.Exec(`
		INSERT INTO trades_q (account, symbol, volume, open, close, side)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"user1", "BTC", 1.5, 100.0, 110.123456, "buy")
	if err != nil {
		t.Fatalf("Не удалось вставить тестовые данные: %v", err)
	}

	err = tradeService.ProcessTrades()
	if err != nil {
		t.Fatalf("Ожидалось успешное выполнение ProcessTrades, но произошла ошибка: %v", err)
	}

	var profit float64
	err = db.QueryRow("SELECT profit FROM account_stats WHERE account = 'user1'").Scan(&profit)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к account_stats: %v", err)
	}
	expectedProfit := 1518518.4
	if math.Abs(profit-expectedProfit) > 0.01 {
		t.Errorf("Ожидалась прибыль %.1f, но найдено: %.1f", expectedProfit, profit)
	}
}

func TestProcessTrades_ErrorInLoop(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tradeService := NewTradeService(db)

	_, err := db.Exec(`
		INSERT INTO trades_q (account, symbol, volume, open, close, side)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"user1", "BTC", 1.5, 100.0, 110.0, "buy")
	if err != nil {
		t.Fatalf("Не удалось вставить тестовые данные: %v", err)
	}

	_, err = db.Exec("DROP TABLE account_stats")
	if err != nil {
		t.Fatalf("Не удалось удалить таблицу account_stats: %v", err)
	}

	err = tradeService.ProcessTrades()
	if err != nil {
		t.Fatalf("Ожидалось успешное выполнение ProcessTrades, несмотря на ошибки в цикле: %v", err)
	}

	var processed int
	err = db.QueryRow("SELECT processed FROM trades_q WHERE account = 'user1'").Scan(&processed)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к trades_q: %v", err)
	}
	if processed != 0 {
		t.Errorf("Ожидалось, что запись не обработана (processed=0), но найдено: %d", processed)
	}
}

func TestProcessTrades_NoRows(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tradeService := NewTradeService(db)

	err := tradeService.ProcessTrades()
	if err != nil {
		t.Fatalf("Ожидалось успешное выполнение ProcessTrades с пустым результатом, но произошла ошибка: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM account_stats").Scan(&count)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к account_stats: %v", err)
	}
	if count != 0 {
		t.Errorf("Ожидалось, что account_stats останется пустым, но найдено: %d", count)
	}
}

func TestProcessTrades_UpdateError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tradeService := NewTradeService(db)

	_, err := db.Exec(`
		INSERT INTO trades_q (account, symbol, volume, open, close, side)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"user1", "BTC", 1.5, 100.0, 110.0, "buy")
	if err != nil {
		t.Fatalf("Не удалось вставить тестовые данные: %v", err)
	}

	_, err = db.Exec("CREATE TRIGGER prevent_update BEFORE UPDATE ON trades_q BEGIN SELECT RAISE(FAIL, 'Update not allowed'); END;")
	if err != nil {
		t.Fatalf("Не удалось создать триггер: %v", err)
	}

	err = tradeService.ProcessTrades()
	if err != nil {
		t.Fatalf("Ожидалось успешное выполнение ProcessTrades несмотря на ошибку обновления: %v", err)
	}

	var processed int
	err = db.QueryRow("SELECT processed FROM trades_q WHERE account = 'user1'").Scan(&processed)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к trades_q: %v", err)
	}
	if processed != 0 {
		t.Errorf("Ожидалось, что запись не обработана (processed=0) из-за ошибки обновления, но найдено: %d", processed)
	}
}

func TestProcessTrades_LogError(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	tradeService := NewTradeService(db)

	_, err := db.Exec(`
		INSERT INTO trades_q (account, symbol, volume, open, close, side)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"user1", "BTC", 1.5, 100.0, 110.0, "invalid_side")
	if err != nil {
		t.Fatalf("Не удалось вставить тестовые данные: %v", err)
	}

	err = tradeService.ProcessTrades()
	if err != nil {
		t.Fatalf("Ожидалась обработка ошибки без возврата: %v", err)
	}

	var processed int
	err = db.QueryRow("SELECT processed FROM trades_q WHERE account = 'user1'").Scan(&processed)
	if err != nil {
		t.Fatalf("Не удалось выполнить запрос к trades_q: %v", err)
	}
	if processed != 0 {
		t.Errorf("Ожидалось, что запись не обработана (processed=0) из-за ошибки, но найдено: %d", processed)
	}
}
