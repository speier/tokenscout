package engine

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/speier/tokenscout/internal/solana"
)

// Monitor watches open positions and triggers exits based on rules
type Monitor struct {
	config        *models.Config
	repo          repository.Repository
	executor      *Executor
	jupiterClient *solana.JupiterClient
}

func NewMonitor(config *models.Config, repo repository.Repository, executor *Executor, jupiterClient *solana.JupiterClient) *Monitor {
	return &Monitor{
		config:        config,
		repo:          repo,
		executor:      executor,
		jupiterClient: jupiterClient,
	}
}

// Start begins monitoring positions
func (m *Monitor) Start(ctx context.Context) error {
	logger.Info().Msg("👀 Starting position monitor")

	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Position monitor shutting down")
			return nil
		case <-ticker.C:
			if err := m.checkPositions(ctx); err != nil {
				logger.Error().Err(err).Msg("Failed to check positions")
			}
		}
	}
}

func (m *Monitor) checkPositions(ctx context.Context) error {
	positions, err := m.repo.GetAllPositions(ctx)
	if err != nil {
		return err
	}

	if len(positions) == 0 {
		return nil
	}

	for _, pos := range positions {
		// Check max trade duration first (always applies)
		duration := time.Since(pos.OpenedAt)
		maxDuration := time.Duration(m.config.Risk.MaxTradeDurationSec) * time.Second

		if duration > maxDuration {
			logger.Info().
				Str("mint", pos.Mint).
				Dur("held", duration).
				Msg("⏱ Max duration exceeded, selling")

			if err := m.executor.ExecuteSell(ctx, pos.Mint, "max_duration_exceeded"); err != nil {
				logger.Error().
					Err(err).
					Str("mint", pos.Mint).
					Msg("Failed to sell position")
			}
			continue
		}

		// Check price-based exits (stop-loss, take-profit)
		if err := m.checkPriceExits(ctx, &pos); err != nil {
			continue
		}
	}

	return nil
}

func (m *Monitor) checkPriceExits(ctx context.Context, pos *models.Position) error {
	// Get current price
	currentPrice, err := m.jupiterClient.GetPrice(ctx, pos.Mint)
	if err != nil {
		return fmt.Errorf("failed to get price: %w", err)
	}

	if currentPrice == 0 {
		return fmt.Errorf("invalid price: 0")
	}

	entryPrice := pos.AvgPriceUSD
	if entryPrice == 0 {
		// No entry price recorded (shouldn't happen, but handle it)
		return fmt.Errorf("no entry price")
	}

	// Parse quantity
	qty, err := strconv.ParseFloat(pos.Quantity, 64)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}

	// Calculate PnL
	_, pnlPct := solana.CalculatePnL(entryPrice, currentPrice, qty)

	// Check take-profit
	if solana.ShouldTakeProfit(entryPrice, currentPrice, m.config.Risk.TakeProfitPct) {
		logger.Info().
			Str("mint", pos.Mint).
			Float64("pnl", pnlPct).
			Msg("📈 Take-profit triggered, selling")

		if err := m.executor.ExecuteSell(ctx, pos.Mint, "take_profit"); err != nil {
			logger.Error().
				Err(err).
				Str("mint", pos.Mint).
				Msg("Failed to sell position")
		}
		return nil
	}

	// Check stop-loss
	if solana.ShouldStopLoss(entryPrice, currentPrice, m.config.Risk.StopLossPct) {
		logger.Info().
			Str("mint", pos.Mint).
			Float64("pnl", pnlPct).
			Msg("📉 Stop-loss triggered, selling")

		if err := m.executor.ExecuteSell(ctx, pos.Mint, "stop_loss"); err != nil {
			logger.Error().
				Err(err).
				Str("mint", pos.Mint).
				Msg("Failed to sell position")
		}
		return nil
	}

	return nil
}
