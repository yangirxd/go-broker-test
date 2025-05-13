package services

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sync"
)

type TradeService struct {
	db *sql.DB
	mu sync.Mutex
}

func NewTradeService(db *sql.DB) *TradeService {
	return &TradeService{db: db}
}

func (s *TradeService) ProcessTrades() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
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

		if side != "buy" && side != "sell" {
			log.Printf("Некорректное значение side: %s для записи с id=%d", side, id)
			continue
		}

		lot := 100000.0
		profit := roundFloat((close-open)*volume*lot, 1)
		if side == "sell" {
			profit = -profit
		}

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

func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
