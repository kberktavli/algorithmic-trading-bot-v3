package domain

import "time"

type Signal struct {
	ID        string    `json:"id"`         // sinyalin unique id
	Symbol    string    `json:"symbol"`     // btcusdt
	Action    string    `json:"action"`     // buy, sell, hold
	Price     float64   `json:"price"`      // işlem anındaki fiyat
	Score     float64   `json:"score"`      // ml botunun predict ettigi güven skoru ( -1,+1 arasında)
	Status    string    `json:"status"`     // pending, completed, failed
	CreatedAt time.Time `json:"created_at"` // sinyalin go servisine ulastıgı zaman
}
