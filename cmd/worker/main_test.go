package main

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/digineat/go-broker-test/internal/db"
)

func setupDB(t *testing.T) *sql.DB {
	// Создаем временную базу данных в памяти
	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось открыть базу данных: %v", err)
	}

	db.InitDB(dbConn)

	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('trades_q', 'account_stats')").Scan(&count)
	if err != nil {
		t.Fatalf("Ошибка при проверке таблиц: %v", err)
	}
	if count != 2 {
		t.Fatalf("Ожидалось 2 таблицы, найдено %d", count)
	}

	return dbConn
}

func setupTestDB(t *testing.T) *sql.DB {
	dbConn, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось открыть базу данных: %v", err)
	}

	db.InitDB(dbConn)

	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('trades_q', 'account_stats')").Scan(&count)
	if err != nil {
		t.Fatalf("Ошибка при проверке таблиц: %v", err)
	}
	if count != 2 {
		t.Fatalf("Ожидалось 2 таблицы, найдено %d", count)
	}

	return dbConn
}

func TestProcessTrades(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*sql.DB) error
		checkResult func(*testing.T, *sql.DB)
		expectError bool
	}{
		{
			name: "успешная обработка записей",
			setup: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO trades_q (account, symbol, volume, open, close, side, processed)
					VALUES
						('ACC1', 'EURUSD', 1.5, 1.2345, 1.2350, 'buy', 0),
						('ACC1', 'GBPUSD', 2.0, 1.4000, 1.3980, 'sell', 0),
						('ACC2', 'USDJPY', 1.0, 110.50, 110.60, 'buy', 0)
				`)
				return err
			},
			checkResult: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM trades_q WHERE processed = 0").Scan(&count)
				if err != nil {
					t.Fatalf("Ошибка при проверке processed: %v", err)
				}
				if count != 0 {
					t.Errorf("Ожидалось 0 необработанных записей, получено %d", count)
				}

				type stat struct {
					account string
					trades  int
					profit  float64
				}
				rows, err := db.Query("SELECT account, trades, profit FROM account_stats")
				if err != nil {
					t.Fatalf("Ошибка при запросе статистики: %v", err)
				}
				defer rows.Close()

				stats := make(map[string]stat)
				for rows.Next() {
					var s stat
					if err := rows.Scan(&s.account, &s.trades, &s.profit); err != nil {
						t.Fatalf("Ошибка при сканировании статистики: %v", err)
					}
					stats[s.account] = s
				}
				acc1, ok := stats["ACC1"]
				if !ok {
					t.Fatal("Статистика для ACC1 не найдена")
				}
				if acc1.trades != 2 {
					t.Errorf("Ожидалось 2 сделки для ACC1, получено %d", acc1.trades)
				}

				expectedProfit := 47.50
				if acc1.profit != expectedProfit {
					t.Errorf("Ожидалась прибыль %.2f для ACC1, получено %.2f", expectedProfit, acc1.profit)
				}

				acc2, ok := stats["ACC2"]
				if !ok {
					t.Fatal("Статистика для ACC2 не найдена")
				}
				if acc2.trades != 1 {
					t.Errorf("Ожидалась 1 сделка для ACC2, получено %d", acc2.trades)
				}
				if acc2.profit != 10.0 {
					t.Errorf("Ожидалась прибыль 10.0 для ACC2, получено %.2f", acc2.profit)
				}
			},
			expectError: false,
		},
		{
			name: "ошибка сканирования (неверный тип данных)",
			setup: func(db *sql.DB) error {
				_, err := db.Exec("INSERT INTO trades_q (account, symbol, volume, open, close, side, processed) VALUES ('ACC3', 'EURUSD', 'invalid', 1.2345, 1.2350, 'buy', 0)")
				return err
			},
			checkResult: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM trades_q WHERE processed = 0").Scan(&count)
				if err != nil {
					t.Fatalf("Ошибка при проверке processed: %v", err)
				}
				if count != 1 {
					t.Errorf("Ожидалась 1 необработанная запись, получено %d", count)
				}
			},
			expectError: false,
		},
		{
			name: "пустая таблица trades_q",
			setup: func(db *sql.DB) error {
				return nil
			},
			checkResult: func(t *testing.T, db *sql.DB) {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM trades_q").Scan(&count)
				if err != nil {
					t.Fatalf("Ошибка при проверке trades_q: %v", err)
				}
				if count != 0 {
					t.Errorf("Ожидалось 0 записей в trades_q, получено %d", count)
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConn := setupDB(t)
			defer dbConn.Close()

			if err := tt.setup(dbConn); err != nil {
				t.Fatalf("Ошибка при настройке теста: %v", err)
			}

			err := processTrades(dbConn)
			if (err != nil) != tt.expectError {
				t.Errorf("Ожидалась ошибка: %v, получено: %v", tt.expectError, err)
			}

			tt.checkResult(t, dbConn)
		})
	}
}

func TestProcessTradesErrors(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(*sql.DB) error
		expectError  bool
		numProcessed int
	}{
		{
			name: "ошибка при запросе",
			setup: func(db *sql.DB) error {
				_, err := db.Exec("DROP TABLE trades_q")
				return err
			},
			expectError:  true,
			numProcessed: 0,
		},
		{
			name: "ошибка при обновлении статистики",
			setup: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO trades_q (account, symbol, volume, open, close, side, processed)
					VALUES ('ACC1', 'EURUSD', 1.0, 1.1000, 1.1050, 'buy', 0)
				`)
				if err != nil {
					return err
				}
				_, err = db.Exec("DROP TABLE account_stats")
				return err
			},
			expectError:  false,
			numProcessed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConn := setupTestDB(t)
			defer dbConn.Close()

			if tt.setup != nil {
				if err := tt.setup(dbConn); err != nil {
					t.Fatalf("Ошибка при настройке теста: %v", err)
				}
			}

			err := processTrades(dbConn)
			if (err != nil) != tt.expectError {
				t.Errorf("processTrades() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.name != "ошибка при запросе" {
				var processed int
				err = dbConn.QueryRow("SELECT COUNT(*) FROM trades_q WHERE processed = 1").Scan(&processed)
				if err != nil {
					t.Logf("Info: could not check processed records: %v", err)
				} else if processed != tt.numProcessed {
					t.Errorf("Expected %d processed records, got %d", tt.numProcessed, processed)
				}
			}
		})
	}
}
