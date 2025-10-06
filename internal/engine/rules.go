package engine

import (
	"context"
	"fmt"

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
		logger.Debug().Err(err).Str("mint", event.Mint).Msg("Failed to fetch token info")
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "failed to fetch token info")
		return decision, nil
	}

	// Check freeze authority
	if r.config.Rules.BlockFreezeAuthority && tokenInfo.HasFreezeAuthority {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "token has freeze authority")
		return decision, nil
	}

	// Check mint authority
	if !r.config.Rules.AllowMintAuthority && tokenInfo.HasMintAuthority {
		decision.Allow = false
		decision.Reasons = append(decision.Reasons, "token has mint authority")
		return decision, nil
	}

	// Check holder count and distribution
	holders, err := solana.GetTokenHolders(ctx, r.rpcClient, event.Mint)
	if err != nil {
		logger.Debug().Err(err).Str("mint", event.Mint).Msg("Failed to fetch holders")
	} else {
		holderCount, topHolderPct, _ := solana.AnalyzeHolderDistribution(holders)
		
		// Check minimum holders
		if holderCount < r.config.Rules.MinHolders {
			decision.Allow = false
			decision.Reasons = append(decision.Reasons, 
				fmt.Sprintf("insufficient holders: %d < %d", holderCount, r.config.Rules.MinHolders))
			return decision, nil
		}

		// Check dev wallet concentration
		if topHolderPct > r.config.Rules.DevWalletMaxPct {
			decision.Allow = false
			decision.Reasons = append(decision.Reasons, 
				fmt.Sprintf("top holder too concentrated: %.2f%% > %.2f%%", topHolderPct, r.config.Rules.DevWalletMaxPct))
			return decision, nil
		}

		logger.Debug().
			Str("mint", event.Mint).
			Int("holders", holderCount).
			Float64("top_holder_pct", topHolderPct).
			Msg("Holder analysis complete")
	}

	// TODO: Check liquidity amount (min_liquidity_usd) - requires DEX pool query
	// TODO: Check token age (max_mint_age_sec) - requires creation timestamp
	// TODO: Simulate sell to detect honeypot

	logger.Debug().
		Str("mint", event.Mint).
		Bool("allow", decision.Allow).
		Strs("reasons", decision.Reasons).
		Msg("Rule evaluation complete")

	if !decision.Allow {
		logger.Info().
			Str("mint", event.Mint).
			Strs("reasons", decision.Reasons).
			Msg("Token rejected by rules")
	}

	return decision, nil
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
