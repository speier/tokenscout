package strategies

import (
	"fmt"
	"github.com/speier/tokenscout/internal/models"
)

// Strategy represents a trading strategy with specific settings
type Strategy struct {
	Name        string
	Description string
	Config      models.Config
}

// Presets contains all built-in strategy presets
var Presets = map[string]Strategy{
	"snipe_flip":        SnipeFlipStrategy(),
	"conservative":      ConservativeStrategy(),
	"scalping":          ScalpingStrategy(),
	"data_collection":   DataCollectionStrategy(),
	"momentum_rider":    MomentumRiderStrategy(),
}

// SnipeFlipStrategy: Catch early pumps, exit fast (3-5 min holds)
// High risk/reward - small positions, high frequency
func SnipeFlipStrategy() Strategy {
	return Strategy{
		Name: "snipe_flip",
		Description: "Snipe & Flip: Catch early pumps, exit fast before rugs (3-5 min holds)",
		Config: models.Config{
			Engine: models.EngineConfig{
				Mode:         models.ModeDryRun,
				MaxPositions: 5,
			},
			Trading: models.TradingConfig{
				BaseMint:                 "SOL",
				QuoteMint:                "USDC",
				MaxSpendPerTrade:         0.2,    // Small positions
				MaxOpenPositions:         5,      // High frequency
				SlippageBps:              400,    // 4% for speed
				PriorityFeeMicroLamports: 20000,  // Fast confirms
			},
			Rules: models.RulesConfig{
				MinLiquidityUSD:      3000,  // Need exit liquidity
				MaxMintAgeSec:        300,   // < 5 min old
				MinHolders:           3,     // Very early entry
				DevWalletMaxPct:      40,    // Safer distribution
				BlockFreezeAuthority: true,  // CRITICAL
				AllowMintAuthority:   false, // CRITICAL
			},
			Risk: models.RiskConfig{
				StopLossPct:         8,   // Quick stop
				TakeProfitPct:       18,  // Fast profits
				MaxTradeDurationSec: 240, // 4 min max
			},
		},
	}
}

// ConservativeStrategy: Safer entry with established tokens (10-20 min holds)
// Lower risk - established tokens, bigger positions, lower frequency
func ConservativeStrategy() Strategy {
	return Strategy{
		Name: "conservative",
		Description: "Conservative: Safer entry with established tokens, longer holds (10-20 min)",
		Config: models.Config{
			Engine: models.EngineConfig{
				Mode:         models.ModeDryRun,
				MaxPositions: 3,
			},
			Trading: models.TradingConfig{
				BaseMint:                 "SOL",
				QuoteMint:                "USDC",
				MaxSpendPerTrade:         0.5,    // Bigger positions
				MaxOpenPositions:         3,      // Lower frequency
				SlippageBps:              200,    // 2% slippage
				PriorityFeeMicroLamports: 10000,  // Normal priority
			},
			Rules: models.RulesConfig{
				MinLiquidityUSD:      10000, // Established pools
				MaxMintAgeSec:        900,   // < 15 min old
				MinHolders:           50,    // More holders
				DevWalletMaxPct:      20,    // Strict distribution
				BlockFreezeAuthority: true,
				AllowMintAuthority:   false,
			},
			Risk: models.RiskConfig{
				StopLossPct:         10,   // Standard stop
				TakeProfitPct:       25,   // Bigger gains
				MaxTradeDurationSec: 900,  // 15 min max
			},
		},
	}
}

// ScalpingStrategy: Ultra-fast flips (30 sec - 2 min holds)
// Highest risk - catch micro pumps, exit immediately
func ScalpingStrategy() Strategy {
	return Strategy{
		Name: "scalping",
		Description: "Scalping: Ultra-fast micro-flips, immediate exits (30 sec - 2 min holds)",
		Config: models.Config{
			Engine: models.EngineConfig{
				Mode:         models.ModeDryRun,
				MaxPositions: 10,
			},
			Trading: models.TradingConfig{
				BaseMint:                 "SOL",
				QuoteMint:                "USDC",
				MaxSpendPerTrade:         0.1,    // Tiny positions
				MaxOpenPositions:         10,     // Very high frequency
				SlippageBps:              500,    // 5% for speed
				PriorityFeeMicroLamports: 50000,  // Highest priority
			},
			Rules: models.RulesConfig{
				MinLiquidityUSD:      2000, // Lower threshold
				MaxMintAgeSec:        180,  // < 3 min old
				MinHolders:           2,    // Ultra early
				DevWalletMaxPct:      50,   // Accept risk
				BlockFreezeAuthority: true,
				AllowMintAuthority:   false,
			},
			Risk: models.RiskConfig{
				StopLossPct:         5,   // Very quick stop
				TakeProfitPct:       10,  // Small quick gains
				MaxTradeDurationSec: 90,  // 1.5 min max
			},
		},
	}
}

// DataCollectionStrategy: Observation mode - no trades, just learn
// Zero risk - tracks all tokens and their outcomes
func DataCollectionStrategy() Strategy {
	return Strategy{
		Name: "data_collection",
		Description: "Data Collection: Observe and learn patterns, no actual trading",
		Config: models.Config{
			Engine: models.EngineConfig{
				Mode:         models.ModeDryRun,
				MaxPositions: 0, // No positions
			},
			Trading: models.TradingConfig{
				BaseMint:                 "SOL",
				QuoteMint:                "USDC",
				MaxSpendPerTrade:         0,      // No spending
				MaxOpenPositions:         0,      // No trades
				SlippageBps:              200,
				PriorityFeeMicroLamports: 5000,
			},
			Rules: models.RulesConfig{
				MinLiquidityUSD:      500,  // Log almost everything
				MaxMintAgeSec:        600,  // < 10 min old
				MinHolders:           1,    // Catch all
				DevWalletMaxPct:      100,  // No filtering
				BlockFreezeAuthority: false, // Log scams too
				AllowMintAuthority:   true,  // Log everything
			},
			Risk: models.RiskConfig{
				StopLossPct:         0,
				TakeProfitPct:       0,
				MaxTradeDurationSec: 3600, // Track for 1 hour
			},
		},
	}
}

// MomentumRiderStrategy: Ride pumps longer (5-15 min holds)
// Medium risk - enter early, hold for bigger gains
func MomentumRiderStrategy() Strategy {
	return Strategy{
		Name: "momentum_rider",
		Description: "Momentum Rider: Enter early with volume, ride pumps for bigger gains (5-15 min)",
		Config: models.Config{
			Engine: models.EngineConfig{
				Mode:         models.ModeDryRun,
				MaxPositions: 4,
			},
			Trading: models.TradingConfig{
				BaseMint:                 "SOL",
				QuoteMint:                "USDC",
				MaxSpendPerTrade:         0.3,    // Medium positions
				MaxOpenPositions:         4,      // Medium frequency
				SlippageBps:              300,    // 3% slippage
				PriorityFeeMicroLamports: 15000,
			},
			Rules: models.RulesConfig{
				MinLiquidityUSD:      5000,  // Good liquidity
				MaxMintAgeSec:        600,   // < 10 min old
				MinHolders:           10,    // Some early holders
				DevWalletMaxPct:      30,    // Medium risk
				BlockFreezeAuthority: true,
				AllowMintAuthority:   false,
			},
			Risk: models.RiskConfig{
				StopLossPct:         15,  // Wider stop
				TakeProfitPct:       40,  // Big gains target
				MaxTradeDurationSec: 600, // 10 min max
			},
		},
	}
}

// GetStrategy returns a strategy by name
func GetStrategy(name string) (Strategy, error) {
	strategy, ok := Presets[name]
	if !ok {
		return Strategy{}, fmt.Errorf("unknown strategy: %s", name)
	}
	return strategy, nil
}

// ListStrategies returns all available strategy names with descriptions
func ListStrategies() []string {
	strategies := []string{}
	for name, strategy := range Presets {
		strategies = append(strategies, fmt.Sprintf("  %-20s - %s", name, strategy.Description))
	}
	return strategies
}

// ApplyStrategy applies a strategy preset to an existing config
// Base config values are overridden by strategy values
func ApplyStrategy(baseConfig *models.Config, strategyName string) (*models.Config, error) {
	strategy, err := GetStrategy(strategyName)
	if err != nil {
		return nil, err
	}

	// Start with strategy config
	config := strategy.Config

	// Preserve some base config values that shouldn't be overridden
	config.Solana = baseConfig.Solana       // Keep RPC/wallet settings
	config.Listener = baseConfig.Listener   // Keep listener settings

	return &config, nil
}
