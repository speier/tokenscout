package models

import "time"

type EventType string

const (
	EventTypeNewMint EventType = "NEW_MINT"
	EventTypeNewPool EventType = "NEW_POOL"
	EventTypeLPAdd   EventType = "LP_ADD"
)

type Event struct {
	ID        int64     `json:"id"`
	Type      EventType `json:"type"`
	Mint      string    `json:"mint"`
	Pair      string    `json:"pair"`
	LPAddress string    `json:"lp_address"`
	Timestamp time.Time `json:"timestamp"`
	Raw       string    `json:"raw"`
}
