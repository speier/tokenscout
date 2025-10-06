package models

// StrategyStats holds performance metrics for a strategy
type StrategyStats struct {
	Strategy       string  `json:"strategy"`
	TotalTrades    int     `json:"total_trades"`
	BuyTrades      int     `json:"buy_trades"`
	SellTrades     int     `json:"sell_trades"`
	OpenPositions  int     `json:"open_positions"`
	AvgEntryPrice  float64 `json:"avg_entry_price"`
	TotalVolume    float64 `json:"total_volume_usd"`
	ExecutedTrades int     `json:"executed_trades"`
	FailedTrades   int     `json:"failed_trades"`
	SuccessRate    float64 `json:"success_rate_pct"`
}
