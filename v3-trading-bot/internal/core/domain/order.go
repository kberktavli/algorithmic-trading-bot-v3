package domain

import "time"

// CCXT ile binance testnet'e iletecegimiz emir yapısı.
type Order struct {
	ID            string    `json:"id"`              // Kendi ürettiğimiz Emir ID'si (UUID vs.)
	SignalID      string    `json:"signal_id"`       // Hangi sinyale dayanarak bu emri açtık?
	Symbol        string    `json:"symbol"`          // Örn: BTC/USDT (CCXT formatında)
	Side          string    `json:"side"`            // "buy" veya "sell"
	OrderType     string    `json:"order_type"`      // "market" veya "limit"
	Amount        float64   `json:"amount"`          // Ne kadar coin alınacak/satılacak?
	Price         float64   `json:"price"`           // Limit emirse fiyat (Market için 0 veya güncel fiyat olabilir)
	ExchangeReqID string    `json:"exchange_req_id"` // Borsanın bize vereceği takip fişi/numarası
	Status        string    `json:"status"`          // "pending", "open", "closed", "canceled", "failed"
	CreatedAt     time.Time `json:"created_at"`
}
