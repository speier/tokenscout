package config

import (
	"fmt"
	"os"

	"github.com/speier/tokenscout/internal/models"
	"github.com/spf13/viper"
)

func Load(configPath string) (*models.Config, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Set defaults
	setDefaults(v)

	// Read config file if it exists
	if _, err := os.Stat(configPath); err == nil {
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg models.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("engine.mode", "dry_run")
	v.SetDefault("engine.max_positions", 3)

	v.SetDefault("solana.rpc_url", "https://api.mainnet-beta.solana.com")
	v.SetDefault("solana.ws_url", "wss://api.mainnet-beta.solana.com")
	v.SetDefault("solana.network", "mainnet-beta")
	v.SetDefault("solana.wallet_path", "wallet.json")
	v.SetDefault("solana.jupiter_api_url", "https://quote-api.jup.ag/v6")

	v.SetDefault("listener.enabled", true)
	v.SetDefault("listener.mode", "websocket") // "websocket" or "polling"
	v.SetDefault("listener.polling_interval_sec", 10) // Poll every 10 seconds
	v.SetDefault("listener.programs", []string{
		"675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8", // Raydium AMM
		"9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP", // Orca Whirlpool
	})
	v.SetDefault("listener.coalesce_window_ms", 200)

	v.SetDefault("trading.base_mint", "SOL")
	v.SetDefault("trading.quote_mint", "USDC")
	v.SetDefault("trading.max_spend_per_trade", 0.5)
	v.SetDefault("trading.max_open_positions", 3)
	v.SetDefault("trading.slippage_bps", 150)
	v.SetDefault("trading.priority_fee_microlamports", 5000)

	v.SetDefault("rules.min_liquidity_usd", 20000)
	v.SetDefault("rules.max_mint_age_sec", 7200)
	v.SetDefault("rules.min_holders", 10) // Lower for fresh tokens
	v.SetDefault("rules.dev_wallet_max_pct", 20)
	v.SetDefault("rules.block_freeze_authority", true)
	v.SetDefault("rules.allow_mint_authority", false)

	v.SetDefault("risk.stop_loss_pct", 10)
	v.SetDefault("risk.take_profit_pct", 10)
	v.SetDefault("risk.max_trade_duration_sec", 600)
}

func CreateDefault(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	v := viper.New()
	setDefaults(v)

	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Create config file
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
