package engine

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/speier/tokenscout/internal/solana"
)

type Decision struct {
	Allow   bool     `json:"allow"`
	Reasons []string `json:"reasons"`
}

type RuleEngine struct {
	config    *models.Config
	repo      repository.Repository
	rpcClient *rpc.Client
}

func NewRuleEngine(config *models.Config, repo repository.Repository, rpcURL string) *RuleEngine {
	return &RuleEngine{
		config:    config,
		repo:      repo,
		rpcClient: rpc.New(rpcURL),
	}
}

func (r *RuleEngine) Evaluate(ctx context.Context, event *models.Event) (*Decision, error) {
	decision := &Decision{
		Allow:   true,
		Reasons: []string{},
	}

	if event.Mint == "" {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "no mint address found")
		return decision, nil
	}

	// Check blacklist
	blacklisted, err := r.repo.IsBlacklisted(ctx, event.Mint)
	if err != nil {
		return nil, fmt.Errorf("failed to check blacklist: %w", err)
	}
	if blacklisted {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "mint is blacklisted")
		return decision, nil
	}

	// Fetch token info
	tokenInfo, err := solana.GetTokenInfo(ctx, r.rpcClient, event.Mint)
	if err != nil {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "failed to fetch token info")
		return decision, nil
	}

	// Check freeze authority
	if r.config.Rules.BlockFreezeAuthority && tokenInfo.HasFreezeAuthority {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "has freeze authority")
		return decision, nil
	}

	// Check mint authority
	if !r.config.Rules.AllowMintAuthority && tokenInfo.HasMintAuthority {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "has mint authority")
		return decision, nil
	}

	// Check holder count and distribution
	holders, err := solana.GetTokenHolders(ctx, r.rpcClient, event.Mint)
	if err != nil {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "failed to fetch holders")
		return decision, nil
	}
	
	holderCount, topHolderPct, _ := solana.AnalyzeHolderDistribution(holders)
	
	// Check minimum holders
	if holderCount < r.config.Rules.MinHolders {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, 
			fmt.Sprintf("holders: %d < %d", holderCount, r.config.Rules.MinHolders))
		return decision, nil
	}

	// Check dev wallet concentration
	if topHolderPct > r.config.Rules.DevWalletMaxPct {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, 
			fmt.Sprintf("top holder: %.1f%% > %.1f%%", topHolderPct, r.config.Rules.DevWalletMaxPct))
		return decision, nil
	}

	// Check token age
	if r.config.Rules.MaxMintAgeSec > 0 {
		tooOld, ageSeconds, err := solana.IsTokenTooOld(ctx, r.rpcClient, event.Mint, int64(r.config.Rules.MaxMintAgeSec))
		if err == nil && tooOld {
			decision.Allow = false
			decision.Reasons = append(decision.Reasons, 
				fmt.Sprintf("too old: %ds", ageSeconds))
			return decision, nil
		}
	}

	// TODO: Check liquidity amount (min_liquidity_usd) - requires DEX pool query
	
	// HONEYPOT DETECTION: Simulate sell to verify token is sellable
	// This is CRITICAL for snipe & flip - many scam tokens allow buy but block sell
	if err := r.checkHoneypot(ctx, event.Mint); err != nil {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, 
			fmt.Sprintf("honeypot detected: %s", err.Error()))
		return decision, nil
	}

	logger.Debug().
		Str("mint", event.Mint).
		Bool("allow", decision.Allow).
		Strs("reasons", decision.Reasons).
		Msg("Rule evaluation complete")

	if !decision.Allow {
		logger.Info().
			Str("mint", formatMint(event.Mint)).
			Str("reason", decision.Reasons[0]).
			Msg("❌ Rejected")
		logger.Debug().
			Strs("all_reasons", decision.Reasons).
			Msg("Full rejection details")
	} else {
		logger.Info().
			Str("mint", formatMint(event.Mint)).
			Msg("✅ Passed all checks")
	}

	return decision, nil
}

// checkHoneypot simulates a sell transaction to verify the token is sellable
// Many scam tokens allow buys but block sells - this catches them before we buy
func (r *RuleEngine) checkHoneypot(ctx context.Context, mint string) error {
	logger.Debug().
		Str("mint", formatMint(mint)).
		Msg("Checking for honeypot (simulating sell)")
	
	// Create Jupiter client for quote
	jupiterClient := solana.NewJupiterClient(r.config.Solana.JupiterAPIURL)
	
	// Simulate selling a small amount: token -> SOL
	// Use a minimal amount (0.001 SOL equivalent) just to test if swap is possible
	testAmount := uint64(1000000) // 0.001 SOL worth of tokens (approximate)
	
	quoteReq := solana.QuoteRequest{
		InputMint:  mint,                                        // Token we want to sell
		OutputMint: "So11111111111111111111111111111111111111112", // SOL (wrapped)
		Amount:     testAmount,
		SlippageBps: 500, // 5% slippage for test
	}
	
	// Try to get a quote for the reverse swap
	quote, err := jupiterClient.GetQuote(ctx, quoteReq)
	if err != nil {
		return fmt.Errorf("cannot get sell quote (likely honeypot)")
	}
	
	if quote == nil || quote.OutAmount == "" {
		return fmt.Errorf("no sell route available (likely honeypot)")
	}
	
	// Parse the output amount to verify it's reasonable
	outAmount, err := strconv.ParseUint(quote.OutAmount, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid sell quote")
	}
	
	// Sanity check: output should be > 0
	if outAmount == 0 {
		return fmt.Errorf("zero output on sell (likely honeypot)")
	}
	
	logger.Debug().
		Str("mint", formatMint(mint)).
		Uint64("out_amount", outAmount).
		Msg("✓ Honeypot check passed (token is sellable)")
	
	return nil
}

// TokenInfo represents metadata about a token
type TokenInfo struct {
	Mint              string
	Name              string
	Symbol            string
	Decimals          uint8
	Supply            uint64
	HasFreezeAuthority bool
	HasMintAuthority   bool
}

// TODO: Implement these helper functions
func (r *RuleEngine) fetchTokenMetadata(ctx context.Context, mint string) (*TokenInfo, error) {
	// TODO: Fetch from Metaplex or direct account queries
	return nil, fmt.Errorf("not implemented")
}

func (r *RuleEngine) checkLiquidity(ctx context.Context, mint string) (float64, error) {
	// TODO: Query DEX pool for liquidity amount
	return 0, fmt.Errorf("not implemented")
}

func (r *RuleEngine) checkHolderCount(ctx context.Context, mint string) (int, error) {
	// TODO: Query token accounts to count holders
	return 0, fmt.Errorf("not implemented")
}

func (r *RuleEngine) checkDevWalletConcentration(ctx context.Context, mint string) (float64, error) {
	// TODO: Analyze top holder wallets
	return 0, fmt.Errorf("not implemented")
}

func (r *RuleEngine) simulateSell(ctx context.Context, mint string, amount uint64) (bool, error) {
	// TODO: Try to get a sell quote - if it fails, might be a honeypot
	return false, fmt.Errorf("not implemented")
}
