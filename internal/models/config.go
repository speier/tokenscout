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
}

type EngineConfig struct {
	Mode         Mode `yaml:"mode"`
	MaxPositions int  `yaml:"max_positions"`
}

type TradingConfig struct {
	BaseMint                  string  `yaml:"base_mint"`
	QuoteMint                 string  `yaml:"quote_mint"`
	MaxSpendPerTrade          float64 `yaml:"max_spend_per_trade"`
	MaxOpenPositions          int     `yaml:"max_open_positions"`
	SlippageBps               int     `yaml:"slippage_bps"`
	PriorityFeeMicroLamports  int64   `yaml:"priority_fee_microlamports"`
}

type RulesConfig struct {
	MinLiquidityUSD      float64 `yaml:"min_liquidity_usd"`
	MaxMintAgeSec        int     `yaml:"max_mint_age_sec"`
	MinHolders           int     `yaml:"min_holders"`
	DevWalletMaxPct      float64 `yaml:"dev_wallet_max_pct"`
	BlockFreezeAuthority bool    `yaml:"block_freeze_authority"`
	AllowMintAuthority   bool    `yaml:"allow_mint_authority"`
}

type SolanaConfig struct {
	RPCURL        string `yaml:"rpc_url" mapstructure:"rpc_url"`
	WSURL         string `yaml:"ws_url" mapstructure:"ws_url"`
	Network       string `yaml:"network" mapstructure:"network"`
	WalletPath    string `yaml:"wallet_path" mapstructure:"wallet_path"`
	JupiterAPIURL string `yaml:"jupiter_api_url" mapstructure:"jupiter_api_url"`
}

type ListenerConfig struct {
	Enabled          bool     `yaml:"enabled"`
	Mode             string   `yaml:"mode"` // "websocket" or "polling"
	PollingInterval  int      `yaml:"polling_interval_sec"` // For polling mode
	Programs         []string `yaml:"programs"`
	CoalesceWindowMs int      `yaml:"coalesce_window_ms"`
}

type RiskConfig struct {
	StopLossPct          float64 `yaml:"stop_loss_pct"`
	TakeProfitPct        float64 `yaml:"take_profit_pct"`
	MaxTradeDurationSec  int     `yaml:"max_trade_duration_sec"`
}
