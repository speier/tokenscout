package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/speier/tokenscout/internal/models"
	"github.com/spf13/viper"
)

var (
	embeddedConfigYAML  string
	embeddedEnvTemplate string
)

// SetEmbeddedTemplates stores the embedded template files
func SetEmbeddedTemplates(configYAML, envTemplate string) {
	embeddedConfigYAML = configYAML
	embeddedEnvTemplate = envTemplate
}

func Load(configPath string) (*models.Config, error) {
	// Load .env file if it exists (silently ignore if not found)
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Enable environment variable support
	v.AutomaticEnv()

	// Bind specific env vars for RPC URLs (priority: .env > config.yaml)
	v.BindEnv("solana.rpc_url", "SOLANA_RPC_URL")
	v.BindEnv("solana.ws_url", "SOLANA_WS_URL")

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

	// Post-process: Inject API key from env into URLs if needed
	if apiKey := os.Getenv("HELIUS_API_KEY"); apiKey != "" {
		cfg.Solana.RPCURL = injectAPIKey(cfg.Solana.RPCURL, apiKey)
		cfg.Solana.WSURL = injectAPIKey(cfg.Solana.WSURL, apiKey)
	}

	return &cfg, nil
}

// injectAPIKey replaces ${HELIUS_API_KEY} placeholder in URL with actual key
func injectAPIKey(url, apiKey string) string {
	return strings.ReplaceAll(url, "${HELIUS_API_KEY}", apiKey)
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("engine.mode", "dry_run")
	v.SetDefault("engine.max_positions", 3)

	// Set trading defaults (must come before engine.max_positions is read elsewhere)
	v.SetDefault("trading.max_open_positions", 3)

	v.SetDefault("solana.rpc_url", "https://api.mainnet-beta.solana.com")
	v.SetDefault("solana.ws_url", "wss://api.mainnet-beta.solana.com")
	v.SetDefault("solana.network", "mainnet-beta")
	v.SetDefault("solana.wallet_path", "wallet.json")
	v.SetDefault("solana.jupiter_api_url", "https://quote-api.jup.ag/v6")

	v.SetDefault("listener.enabled", true)
	v.SetDefault("listener.mode", "websocket")        // "websocket" or "polling"
	v.SetDefault("listener.polling_interval_sec", 10) // Poll every 10 seconds
	// DEX programs to monitor for new token pools (90%+ of new tokens launch here)
	v.SetDefault("listener.programs", []string{
		"675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8", // Raydium AMM (largest DEX)
		"9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP", // Orca Whirlpool (2nd largest DEX)
	})
	v.SetDefault("listener.coalesce_window_ms", 200)

	v.SetDefault("trading.base_mint", "SOL")
	v.SetDefault("trading.quote_mint", "USDC")
	v.SetDefault("trading.max_spend_per_trade", 0.2)          // Smaller positions for high frequency
	v.SetDefault("trading.max_open_positions", 5)             // More concurrent trades
	v.SetDefault("trading.slippage_bps", 400)                 // Higher slippage for speed
	v.SetDefault("trading.priority_fee_microlamports", 20000) // Higher priority for faster confirms

	// Rules tuned for snipe & flip strategy: catch early, exit fast
	v.SetDefault("rules.min_liquidity_usd", 3000)      // Need enough liquidity to exit
	v.SetDefault("rules.max_mint_age_sec", 300)        // Only tokens < 5 minutes old
	v.SetDefault("rules.min_holders", 3)               // Very early entry
	v.SetDefault("rules.dev_wallet_max_pct", 40)       // Safer distribution
	v.SetDefault("rules.block_freeze_authority", true) // CRITICAL: reject if token can be frozen
	v.SetDefault("rules.allow_mint_authority", false)  // CRITICAL: reject if supply can be minted

	// Risk settings for snipe & flip: quick exits
	v.SetDefault("risk.stop_loss_pct", 8)            // Quick exit on loss
	v.SetDefault("risk.take_profit_pct", 18)         // Take profits fast (don't be greedy)
	v.SetDefault("risk.max_trade_duration_sec", 240) // 4 min max hold (exit before rugs)
}

func CreateDefault(configPath string) error {
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	// Use embedded template if available, otherwise generate from defaults
	if embeddedConfigYAML != "" {
		if err := os.WriteFile(configPath, []byte(embeddedConfigYAML), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	} else {
		// Fallback: generate from viper defaults
		v := viper.New()
		setDefaults(v)
		v.SetConfigFile(configPath)
		v.SetConfigType("yaml")
		if err := v.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	return nil
}

func CreateEnvTemplate(envPath string) error {
	if _, err := os.Stat(envPath); err == nil {
		return fmt.Errorf("env file already exists: %s", envPath)
	}

	if embeddedEnvTemplate == "" {
		return fmt.Errorf("embedded env template not available")
	}

	if err := os.WriteFile(envPath, []byte(embeddedEnvTemplate), 0600); err != nil {
		return fmt.Errorf("failed to write env file: %w", err)
	}

	return nil
}
