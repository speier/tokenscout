package models

import "time"

type TradeSide string

const (
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

type TradeStatus string

const (
	TradeStatusPending  TradeStatus = "PENDING"
	TradeStatusExecuted TradeStatus = "EXECUTED"
	TradeStatusFailed   TradeStatus = "FAILED"
)

type Trade struct {
	ID        int64       `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Side      TradeSide   `json:"side"`
	Mint      string      `json:"mint"`
	Quantity  string      `json:"quantity"`
	PriceUSD  float64     `json:"price_usd"`
	TxSig     string      `json:"tx_sig"`
	Status    TradeStatus `json:"status"`
	Strategy  string      `json:"strategy"` // Strategy name used for this trade
}
