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

func formatMint(mint string) string {
	if len(mint) > 8 {
		return mint[:4] + ".." + mint[len(mint)-4:]
	}
	return mint
}

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
		Str("mint", formatMint(mint)).
		Msg("💰 Preparing to buy...")
	logger.Debug().
		Str("mode", string(e.config.Engine.Mode)).
		Float64("amount", e.config.Trading.MaxSpendPerTrade).
		Msg("Buy order details")

	// Create trade record
	trade := &models.Trade{
		Timestamp: time.Now(),
		Side:      models.TradeSideBuy,
		Mint:      mint,
		Quantity:  fmt.Sprintf("%.9f", e.config.Trading.MaxSpendPerTrade),
		Status:    models.TradeStatusPending,
		Strategy:  e.config.Strategy,
	}

	if err := e.repo.CreateTrade(ctx, trade); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	// Get real Jupiter quote (both dry-run and live modes)
	jupiterClient := e.jupiterClient

	// Convert SOL amount to lamports
	solAmount := uint64(e.config.Trading.MaxSpendPerTrade * 1e9)

	quoteReq := solana.QuoteRequest{
		InputMint:        "So11111111111111111111111111111111111111112", // SOL
		OutputMint:       mint,
		Amount:           solAmount,
		SlippageBps:      100, // 1% slippage
		OnlyDirectRoutes: false,
	}

	logger.Debug().
		Str("mint", mint).
		Uint64("sol_amount", solAmount).
		Msg("Fetching Jupiter quote")

	quote, err := jupiterClient.GetQuote(ctx, quoteReq)
	if err != nil {
		logger.Error().Err(err).Str("mint", mint).Msg("Failed to get Jupiter quote")
		if err := e.repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusFailed, err.Error()); err != nil {
			return fmt.Errorf("failed to update trade: %w", err)
		}
		return fmt.Errorf("failed to get quote: %w", err)
	}

	// Parse amounts to calculate real price
	inAmountFloat, err := strconv.ParseFloat(quote.InAmount, 64)
	if err != nil {
		return fmt.Errorf("failed to parse input amount: %w", err)
	}
	outAmountFloat, err := strconv.ParseFloat(quote.OutAmount, 64)
	if err != nil {
		return fmt.Errorf("failed to parse output amount: %w", err)
	}

	// Calculate token quantity (accounting for decimals)
	tokenQuantity := outAmountFloat / 1e9 // Assuming 9 decimals, should fetch from token mint

	// Get SOL/USD price to calculate token price in USD
	solPrice, err := solana.GetSOLPrice(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get SOL price, using $100 fallback")
		solPrice = 100.0
	}

	// Calculate real token price in USD
	solSpent := inAmountFloat / 1e9
	usdSpent := solSpent * solPrice
	tokenPriceUSD := usdSpent / tokenQuantity

	logger.Info().
		Str("mint", formatMint(mint)).
		Float64("price_usd", tokenPriceUSD).
		Float64("tokens", tokenQuantity).
		Msg("💵 Quote received")
	logger.Debug().
		Float64("sol_spent", solSpent).
		Float64("price_impact_pct", parsePriceImpact(quote.PriceImpactPct)).
		Msg("Quote details")

	// In dry-run mode, don't execute but use real prices
	if e.config.Engine.Mode == models.ModeDryRun {
		logger.Info().
			Str("mint", formatMint(mint)).
			Msg("✅ Simulated buy (dry-run mode)")
		logger.Debug().Msg("Would execute actual swap on Jupiter")

		// Update trade as executed (simulated)
		if err := e.repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusExecuted, "DRY_RUN"); err != nil {
			return fmt.Errorf("failed to update trade: %w", err)
		}

		// Create position with REAL price from Jupiter quote
		position := &models.Position{
			Mint:         mint,
			Quantity:     fmt.Sprintf("%.9f", tokenQuantity),
			AvgPriceUSD:  tokenPriceUSD, // REAL price from quote
			OpenedAt:     time.Now(),
			LastUpdateAt: time.Now(),
			Strategy:     e.config.Strategy,
		}

		if err := e.repo.CreatePosition(ctx, position); err != nil {
			return fmt.Errorf("failed to create position: %w", err)
		}

		logger.Info().
			Str("mint", formatMint(mint)).
			Float64("entry_price", tokenPriceUSD).
			Msg("📈 Position opened")

		return nil
	}

	// Live mode - execute actual trade with the quote we already have
	logger.Info().
		Str("mint", mint).
		Msg("LIVE: Executing real buy via Jupiter")

	// TODO: Execute swap transaction
	// swapReq := solana.SwapRequest{
	// 	QuoteResponse:    *quote,
	// 	UserPublicKey:    wallet.PublicKey().String(),
	// 	WrapAndUnwrapSol: true,
	// }
	//
	// swapResp, err := jupiterClient.GetSwap(ctx, swapReq)
	// ... sign and send transaction ...

	logger.Warn().Msg("LIVE: Swap execution not yet implemented")
	return fmt.Errorf("live trading not yet implemented")
}

// Helper to parse price impact percentage
func parsePriceImpact(impact string) float64 {
	val, err := strconv.ParseFloat(impact, 64)
	if err != nil {
		return 0.0
	}
	return val
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
		Strategy:  position.Strategy, // Use strategy from the position
	}

	if err := e.repo.CreateTrade(ctx, trade); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	// Get real Jupiter quote for selling
	jupiterClient := e.jupiterClient

	// Parse token quantity
	tokenAmount, err := strconv.ParseFloat(position.Quantity, 64)
	if err != nil {
		return fmt.Errorf("failed to parse quantity: %w", err)
	}

	// Convert to token units (assuming 9 decimals)
	tokenUnits := uint64(tokenAmount * 1e9)

	quoteReq := solana.QuoteRequest{
		InputMint:        mint,
		OutputMint:       "So11111111111111111111111111111111111111112", // SOL
		Amount:           tokenUnits,
		SlippageBps:      100, // 1% slippage
		OnlyDirectRoutes: false,
	}

	logger.Debug().
		Str("mint", mint).
		Uint64("token_amount", tokenUnits).
		Msg("Fetching Jupiter sell quote")

	quote, err := jupiterClient.GetQuote(ctx, quoteReq)
	if err != nil {
		logger.Error().Err(err).Str("mint", mint).Msg("Failed to get Jupiter sell quote")
		if err := e.repo.UpdateTradeStatus(ctx, trade.ID, models.TradeStatusFailed, err.Error()); err != nil {
			return fmt.Errorf("failed to update trade: %w", err)
		}
		return fmt.Errorf("failed to get sell quote: %w", err)
	}

	// Calculate SOL received
	outAmountFloat, err := strconv.ParseFloat(quote.OutAmount, 64)
	if err != nil {
		return fmt.Errorf("failed to parse output amount: %w", err)
	}
	solReceived := outAmountFloat / 1e9

	// Get SOL price
	solPrice, err := solana.GetSOLPrice(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to get SOL price, using $100 fallback")
		solPrice = 100.0
	}

	usdReceived := solReceived * solPrice

	logger.Info().
		Str("mint", mint).
		Float64("sol_received", solReceived).
		Float64("usd_received", usdReceived).
		Float64("price_impact", parsePriceImpact(quote.PriceImpactPct)).
		Msg("Real sell quote received from Jupiter")

	// In dry-run mode, don't execute but use real prices
	if e.config.Engine.Mode == models.ModeDryRun {
		logger.Info().
			Str("mint", mint).
			Str("reason", reason).
			Msg("DRY RUN: Would execute sell via Jupiter (using real quote)")

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
			Float64("usd_received", usdReceived).
			Msg("DRY RUN: Position closed with real sell price")

		return nil
	}

	// Live mode - execute actual trade
	logger.Info().
		Str("mint", mint).
		Msg("LIVE: Executing real sell via Jupiter")

	// TODO: Execute swap transaction
	logger.Warn().Msg("LIVE: Sell swap execution not yet implemented")
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
