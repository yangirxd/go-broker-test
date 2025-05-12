package db

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func TestInitDB(t *testing.T) {
	// Создаем временную базу данных в памяти
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Не удалось открыть базу данных: %v", err)
	}
	defer db.Close()

	InitDB(db)

	tests := []struct {
		name      string
		query     string
		checkFunc func(*testing.T, *sql.Rows)
	}{
		{
			name:  "проверка создания таблицы trades_q",
			query: "SELECT name, sql FROM sqlite_master WHERE type='table' AND name='trades_q'",
			checkFunc: func(t *testing.T, rows *sql.Rows) {
				if !rows.Next() {
					t.Fatal("Таблица trades_q не создана")
				}
				var name, sql string
				if err := rows.Scan(&name, &sql); err != nil {
					t.Fatalf("Ошибка при чтении результата: %v", err)
				}
				if name != "trades_q" {
					t.Errorf("Ожидалось имя таблицы trades_q, получено %s", name)
				}
				sqlLower := strings.ToLower(sql)
				expectedColumns := []string{
					"id integer primary key autoincrement",
					"account text not null",
					"symbol text not null",
					"volume real not null",
					"open real not null",
					"close real not null",
					"side text not null",
					"processed integer default 0",
				}
				for _, col := range expectedColumns {
					if !containsIgnoreCase(sqlLower, col) {
						t.Errorf("Ожидался столбец %s в определении таблицы, но не найден", col)
					}
				}
			},
		},
		{
			name:  "проверка создания таблицы account_stats",
			query: "SELECT name, sql FROM sqlite_master WHERE type='table' AND name='account_stats'",
			checkFunc: func(t *testing.T, rows *sql.Rows) {
				if !rows.Next() {
					t.Fatal("Таблица account_stats не создана")
				}
				var name, sql string
				if err := rows.Scan(&name, &sql); err != nil {
					t.Fatalf("Ошибка при чтении результата: %v", err)
				}
				if name != "account_stats" {
					t.Errorf("Ожидалось имя таблицы account_stats, получено %s", name)
				}
				sqlLower := strings.ToLower(sql)
				expectedColumns := []string{
					"account text primary key",
					"trades integer default 0",
					"profit real default 0.0",
				}
				for _, col := range expectedColumns {
					if !containsIgnoreCase(sqlLower, col) {
						t.Errorf("Ожидался столбец %s в определении таблицы, но не найден", col)
					}
				}
			},
		},
		{
			name:  "проверка PRAGMA journal_mode",
			query: "PRAGMA journal_mode",
			checkFunc: func(t *testing.T, rows *sql.Rows) {
				if !rows.Next() {
					t.Fatal("Не удалось получить journal_mode")
				}
				var mode string
				if err := rows.Scan(&mode); err != nil {
					t.Fatalf("Ошибка при чтении journal_mode: %v", err)
				}
				mode = strings.ToLower(mode)
				if mode != "wal" && mode != "memory" {
					t.Errorf("Ожидался journal_mode=wal или memory, получено %s", mode)
				}
			},
		},
		{
			name:  "проверка PRAGMA busy_timeout",
			query: "PRAGMA busy_timeout",
			checkFunc: func(t *testing.T, rows *sql.Rows) {
				if !rows.Next() {
					t.Fatal("Не удалось получить busy_timeout")
				}
				var timeout int
				if err := rows.Scan(&timeout); err != nil {
					t.Fatalf("Ошибка при чтении busy_timeout: %v", err)
				}
				if timeout != 5000 {
					t.Errorf("Ожидался busy_timeout=5000, получено %d", timeout)
				}
			},
		},
		{
			name:  "проверка PRAGMA synchronous",
			query: "PRAGMA synchronous",
			checkFunc: func(t *testing.T, rows *sql.Rows) {
				if !rows.Next() {
					t.Fatal("Не удалось получить synchronous")
				}
				var sync int
				if err := rows.Scan(&sync); err != nil {
					t.Fatalf("Ошибка при чтении synchronous: %v", err)
				}
				if sync != 1 {
					t.Errorf("Ожидался synchronous=NORMAL (1), получено %d", sync)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, err := db.Query(tt.query)
			if err != nil {
				t.Fatalf("Ошибка выполнения запроса %s: %v", tt.query, err)
			}
			defer rows.Close()
			tt.checkFunc(t, rows)
		})
	}
}
