package models

import "time"

type Position struct {
	Mint         string    `json:"mint"`
	Quantity     string    `json:"quantity"`
	AvgPriceUSD  float64   `json:"avg_price_usd"`
	OpenedAt     time.Time `json:"opened_at"`
	LastUpdateAt time.Time `json:"last_update_at"`
}
