package repository

import (
	"context"
	"github.com/speier/tokenscout/internal/models"
)

type Repository interface {
	// Trades
	CreateTrade(ctx context.Context, trade *models.Trade) error
	GetTrades(ctx context.Context, limit int) ([]models.Trade, error)
	GetTradeByID(ctx context.Context, id int64) (*models.Trade, error)
	UpdateTradeStatus(ctx context.Context, id int64, status models.TradeStatus, txSig string) error

	// Positions
	CreatePosition(ctx context.Context, position *models.Position) error
	GetPosition(ctx context.Context, mint string) (*models.Position, error)
	GetAllPositions(ctx context.Context) ([]models.Position, error)
	UpdatePosition(ctx context.Context, position *models.Position) error
	DeletePosition(ctx context.Context, mint string) error

	// Events
	CreateEvent(ctx context.Context, event *models.Event) error
	GetRecentEvents(ctx context.Context, limit int) ([]models.Event, error)

	// Config
	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key, value string) error

	// Blacklist/Whitelist
	IsBlacklisted(ctx context.Context, mint string) (bool, error)
	IsWhitelisted(ctx context.Context, mint string) (bool, error)
	AddToBlacklist(ctx context.Context, mint string) error
	AddToWhitelist(ctx context.Context, mint string) error

	// Utility
	Close() error
}
