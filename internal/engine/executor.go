package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/speier/tokenscout/internal/solana"
)

type Executor struct {
	config        *models.Config
	repo          repository.Repository
	solanaClient  *solana.Client
	jupiterClient *solana.JupiterClient
}

func NewExecutor(
	config *models.Config,
	repo repository.Repository,
	solanaClient *solana.Client,
	jupiterClient *solana.JupiterClient,
) *Executor {
	return &Executor{
		config:        config,
		repo:          repo,
		solanaClient:  solanaClient,
		jupiterClient: jupiterClient,
	}
}

// ExecuteBuy opens a new position by buying a token
func (e *Executor) ExecuteBuy(ctx context.Context, mint string, reason string) error {
	// Check if already have a position
	existingPos, err := e.repo.GetPosition(ctx, mint)
	if err == nil && existingPos != nil {
		logger.Info().Str("mint", mint).Msg("Position already exists, skipping buy")
		return nil
	}

	// Check if we can open more positions
	positions, err := e.repo.GetAllPositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	if len(positions) >= e.config.Trading.MaxOpenPositions {
		logger.Info().
			Int("current", len(positions)).
			Int("max", e.config.Trading.MaxOpenPositions).
			Msg("Max positions reached, skipping buy")
		return nil
	}

	logger.Info().
		Str("mint", mint).
		Str("mode", string(e.config.Engine.Mode)).
		Float64("amount", e.config.Trading.MaxSpendPerTrade).
		Msg("Executing BUY order")

	// Create trade record
	trade := &models.Trade{
		Timestamp: time.Now(),
		Side:      models.TradeSideBuy,
		Mint:      mint,
		Quantity:  fmt.Sprintf("%.9f", e.config.Trading.MaxSpendPerTrade),
		Status:    models.TradeStatusPending,
	}

	if err := e.repo.CreateTrade(ctx, trade); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	// In dry-run mode, just simulate
	if e.config.Engine.Mode == models.ModeDryRun {
		logger.Info().
			Str("mint", mint).
			Msg("DRY RUN: Would execute buy via Jupiter")
		
		// Update trade as executed (simulated)
		if err := e.repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusExecuted, "DRY_RUN"); err != nil {
			return fmt.Errorf("failed to update trade: %w", err)
		}

		// Create simulated position
		position := &models.Position{
			Mint:         mint,
			Quantity:     trade.Quantity,
			AvgPriceUSD:  100.0, // Simulated price
			OpenedAt:     time.Now(),
			LastUpdateAt: time.Now(),
		}

		if err := e.repo.CreatePosition(ctx, position); err != nil {
			return fmt.Errorf("failed to create position: %w", err)
		}

		logger.Info().
			Str("mint", mint).
			Msg("DRY RUN: Position opened (simulated)")
		
		return nil
	}

	// Live mode - execute actual trade
	// TODO: Get actual quote and execute swap
	logger.Info().
		Str("mint", mint).
		Msg("LIVE: Would execute real buy via Jupiter (not yet implemented)")

	return fmt.Errorf("live trading not yet implemented")
}

// ExecuteSell closes a position by selling a token
func (e *Executor) ExecuteSell(ctx context.Context, mint string, reason string) error {
	// Get position
	position, err := e.repo.GetPosition(ctx, mint)
	if err != nil {
		return fmt.Errorf("position not found: %w", err)
	}

	logger.Info().
		Str("mint", mint).
		Str("reason", reason).
		Str("mode", string(e.config.Engine.Mode)).
		Msg("Executing SELL order")

	// Create trade record
	trade := &models.Trade{
		Timestamp: time.Now(),
		Side:      models.TradeSideSell,
		Mint:      mint,
		Quantity:  position.Quantity,
		Status:    models.TradeStatusPending,
	}

	if err := e.repo.CreateTrade(ctx, trade); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	// In dry-run mode, just simulate
	if e.config.Engine.Mode == models.ModeDryRun {
		logger.Info().
			Str("mint", mint).
			Str("reason", reason).
			Msg("DRY RUN: Would execute sell via Jupiter")
		
		// Update trade as executed (simulated)
		if err := e.repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusExecuted, "DRY_RUN"); err != nil {
			return fmt.Errorf("failed to update trade: %w", err)
		}

		// Delete position
		if err := e.repo.DeletePosition(ctx, mint); err != nil {
			return fmt.Errorf("failed to delete position: %w", err)
		}

		logger.Info().
			Str("mint", mint).
			Msg("DRY RUN: Position closed (simulated)")
		
		return nil
	}

	// Live mode - execute actual trade
	logger.Info().
		Str("mint", mint).
		Msg("LIVE: Would execute real sell via Jupiter (not yet implemented)")

	return fmt.Errorf("live trading not yet implemented")
}

// SellAll closes all open positions
func (e *Executor) SellAll(ctx context.Context, reason string) error {
	positions, err := e.repo.GetAllPositions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get positions: %w", err)
	}

	logger.Info().
		Int("count", len(positions)).
		Str("reason", reason).
		Msg("Selling all positions")

	for _, pos := range positions {
		if err := e.ExecuteSell(ctx, pos.Mint, reason); err != nil {
			logger.Error().
				Err(err).
				Str("mint", pos.Mint).
				Msg("Failed to sell position")
		}
	}

	return nil
}
