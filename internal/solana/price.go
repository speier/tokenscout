package solana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PriceInfo represents token price information
type PriceInfo struct {
	Mint       string
	PriceUSD   float64
	PriceSOL   float64
	Liquidity  float64
	Source     string // "raydium", "orca", "jupiter"
}

// GetTokenPrice fetches current token price
// This queries Jupiter's price API for real-time pricing
func GetTokenPrice(ctx context.Context, jupiterClient *JupiterClient, mint string) (*PriceInfo, error) {
	// TODO: Implement price fetching
	// Options:
	// 1. Jupiter price API: https://price.jup.ag/v4/price?ids=MINT
	// 2. Get quote from Jupiter (1 token → USDC)
	// 3. Query pool reserves directly from Raydium/Orca
	
	// For now, return mock data
	return &PriceInfo{
		Mint:     mint,
		PriceUSD: 0.001, // Placeholder
		Source:   "mock",
	}, nil
}

// CalculatePnL calculates profit/loss for a position
func CalculatePnL(entryPrice, currentPrice, quantity float64) (pnl float64, pnlPct float64) {
	if entryPrice == 0 {
		return 0, 0
	}
	
	pnl = (currentPrice - entryPrice) * quantity
	pnlPct = ((currentPrice - entryPrice) / entryPrice) * 100
	
	return pnl, pnlPct
}

// ShouldTakeProfit checks if take-profit threshold is reached
func ShouldTakeProfit(entryPrice, currentPrice, takeProfitPct float64) bool {
	if entryPrice == 0 {
		return false
	}
	_, pnlPct := CalculatePnL(entryPrice, currentPrice, 1)
	return pnlPct >= takeProfitPct
}

// ShouldStopLoss checks if stop-loss threshold is reached
func ShouldStopLoss(entryPrice, currentPrice, stopLossPct float64) bool {
	if entryPrice == 0 {
		return false
	}
	_, pnlPct := CalculatePnL(entryPrice, currentPrice, 1)
	return pnlPct <= -stopLossPct
}

// JupiterPriceResponse represents Jupiter Price API response
type JupiterPriceResponse struct {
	Data map[string]struct {
		ID        string  `json:"id"`
		MintSymbol string `json:"mintSymbol"`
		VsToken   string  `json:"vsToken"`
		VsTokenSymbol string `json:"vsTokenSymbol"`
		Price     float64 `json:"price"`
	} `json:"data"`
	TimeTaken float64 `json:"timeTaken"`
}

// GetPrice fetches price from Jupiter Price API
func (j *JupiterClient) GetPrice(ctx context.Context, mint string) (float64, error) {
	// Try Jupiter Price API first
	priceURL := fmt.Sprintf("https://price.jup.ag/v4/price?ids=%s", mint)
	
	req, err := http.NewRequestWithContext(ctx, "GET", priceURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to quote method
		return j.getPriceViaQuote(ctx, mint)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return j.getPriceViaQuote(ctx, mint)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return j.getPriceViaQuote(ctx, mint)
	}
	
	var priceResp JupiterPriceResponse
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return j.getPriceViaQuote(ctx, mint)
	}
	
	if data, ok := priceResp.Data[mint]; ok {
		// Price is in USD already
		return data.Price, nil
	}
	
	// Fallback to quote
	return j.getPriceViaQuote(ctx, mint)
}

// getPriceViaQuote estimates price using Jupiter quote (fallback)
func (j *JupiterClient) getPriceViaQuote(ctx context.Context, mint string) (float64, error) {
	baseMint := "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // USDC
	
	quote, err := j.GetQuote(ctx, QuoteRequest{
		InputMint:   mint,
		OutputMint:  baseMint,
		Amount:      1_000_000, // 1 token (6 decimals)
		SlippageBps: 100,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get quote: %w", err)
	}
	
	var inAmount, outAmount uint64
	fmt.Sscanf(quote.InAmount, "%d", &inAmount)
	fmt.Sscanf(quote.OutAmount, "%d", &outAmount)
	
	if inAmount == 0 {
		return 0, fmt.Errorf("invalid quote amounts")
	}
	
	priceUSD := float64(outAmount) / float64(inAmount)
	return priceUSD, nil
}

// GetPoolLiquidity estimates pool liquidity for a token
func GetPoolLiquidity(ctx context.Context, mint string) (float64, error) {
	// TODO: Query Raydium/Orca pool accounts
	// Parse pool reserves
	// Calculate total liquidity in USD
	return 0, fmt.Errorf("not implemented")
}

// ParsePoolReserves parses pool account data to get reserves
func ParsePoolReserves(data []byte) (reserve0, reserve1 uint64, err error) {
	// TODO: Parse Raydium or Orca pool data structure
	return 0, 0, fmt.Errorf("not implemented")
}

// GetSOLPrice fetches current SOL/USD price from Jupiter
func GetSOLPrice(ctx context.Context) (float64, error) {
	priceURL := "https://price.jup.ag/v4/price?ids=So11111111111111111111111111111111111111112"
	
	req, err := http.NewRequestWithContext(ctx, "GET", priceURL, nil)
	if err != nil {
		return 100.0, fmt.Errorf("failed to create request: %w", err)
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Return approximate fallback
		return 100.0, nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 100.0, nil
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 100.0, nil
	}
	
	var priceResp JupiterPriceResponse
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return 100.0, nil
	}
	
	if data, ok := priceResp.Data["So11111111111111111111111111111111111111112"]; ok {
		return data.Price, nil
	}
	
	return 100.0, nil
}

// ConvertSOLToUSD converts SOL amount to USD using current price
func ConvertSOLToUSD(solAmount float64) float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	solPrice, err := GetSOLPrice(ctx)
	if err != nil {
		solPrice = 100.0 // Fallback
	}
	return solAmount * solPrice
}

// ConvertLamportsToSOL converts lamports to SOL
func ConvertLamportsToSOL(lamports uint64) float64 {
	return float64(lamports) / 1e9
}

// ConvertSOLToLamports converts SOL to lamports
func ConvertSOLToLamports(sol float64) uint64 {
	return uint64(sol * 1e9)
}

// EstimateSwapOutput estimates output amount for a swap
func EstimateSwapOutput(inputAmount, reserve0, reserve1 uint64) uint64 {
	// Constant product AMM formula: x * y = k
	// Output = (inputAmount * reserve1) / (reserve0 + inputAmount)
	if reserve0 == 0 {
		return 0
	}
	return (inputAmount * reserve1) / (reserve0 + inputAmount)
}
