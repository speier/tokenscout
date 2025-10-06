package models

type Mode string

const (
	ModeLive   Mode = "live"
	ModeDryRun Mode = "dry_run"
)

type Config struct {
	Engine   EngineConfig   `yaml:"engine"`
	Solana   SolanaConfig   `yaml:"solana"`
	Listener ListenerConfig `yaml:"listener"`
	Trading  TradingConfig  `yaml:"trading"`
	Rules    RulesConfig    `yaml:"rules"`
	Risk     RiskConfig     `yaml:"risk"`
	Strategy string         `yaml:"-"` // Runtime strategy name, not persisted in config.yaml
}

type EngineConfig struct {
	Mode         Mode `yaml:"mode" mapstructure:"mode"`
	MaxPositions int  `yaml:"max_positions" mapstructure:"max_positions"`
}

type TradingConfig struct {
	BaseMint                 string  `yaml:"base_mint" mapstructure:"base_mint"`
	QuoteMint                string  `yaml:"quote_mint" mapstructure:"quote_mint"`
	MaxSpendPerTrade         float64 `yaml:"max_spend_per_trade" mapstructure:"max_spend_per_trade"`
	MaxOpenPositions         int     `yaml:"max_open_positions" mapstructure:"max_open_positions"`
	SlippageBps              int     `yaml:"slippage_bps" mapstructure:"slippage_bps"`
	PriorityFeeMicroLamports int64   `yaml:"priority_fee_microlamports" mapstructure:"priority_fee_microlamports"`
}

type RulesConfig struct {
	MinLiquidityUSD      float64 `yaml:"min_liquidity_usd" mapstructure:"min_liquidity_usd"`
	MaxMintAgeSec        int     `yaml:"max_mint_age_sec" mapstructure:"max_mint_age_sec"`
	MinHolders           int     `yaml:"min_holders" mapstructure:"min_holders"`
	DevWalletMaxPct      float64 `yaml:"dev_wallet_max_pct" mapstructure:"dev_wallet_max_pct"`
	BlockFreezeAuthority bool    `yaml:"block_freeze_authority" mapstructure:"block_freeze_authority"`
	AllowMintAuthority   bool    `yaml:"allow_mint_authority" mapstructure:"allow_mint_authority"`
}

type SolanaConfig struct {
	RPCURL        string `yaml:"rpc_url" mapstructure:"rpc_url"`
	WSURL         string `yaml:"ws_url" mapstructure:"ws_url"`
	Network       string `yaml:"network" mapstructure:"network"`
	WalletPath    string `yaml:"wallet_path" mapstructure:"wallet_path"`
	JupiterAPIURL string `yaml:"jupiter_api_url" mapstructure:"jupiter_api_url"`
}

type ListenerConfig struct {
	Enabled          bool     `yaml:"enabled" mapstructure:"enabled"`
	Mode             string   `yaml:"mode" mapstructure:"mode"`                                 // "webhook", "websocket", or "polling"
	PollingInterval  int      `yaml:"polling_interval_sec" mapstructure:"polling_interval_sec"` // For polling mode
	Programs         []string `yaml:"programs" mapstructure:"programs"`
	CoalesceWindowMs int      `yaml:"coalesce_window_ms" mapstructure:"coalesce_window_ms"`
	WebhookPort      int      `yaml:"webhook_port" mapstructure:"webhook_port"`     // Port for webhook server
	WebhookPath      string   `yaml:"webhook_path" mapstructure:"webhook_path"`     // Path for webhook endpoint
	WebhookSecret    string   `yaml:"webhook_secret" mapstructure:"webhook_secret"` // Optional: verify webhook requests
}

type RiskConfig struct {
	StopLossPct         float64 `yaml:"stop_loss_pct" mapstructure:"stop_loss_pct"`
	TakeProfitPct       float64 `yaml:"take_profit_pct" mapstructure:"take_profit_pct"`
	MaxTradeDurationSec int     `yaml:"max_trade_duration_sec" mapstructure:"max_trade_duration_sec"`
}
