# Setup Guide

## Installation

### Option 1: Pre-built Binary
Download from [Releases](https://github.com/speier/tokenscout/releases):

**macOS:**
```bash
curl -LO https://github.com/speier/tokenscout/releases/latest/download/tokenscout_*_Darwin_arm64.tar.gz
tar -xzf tokenscout_*.tar.gz
./tokenscout version
```

**Linux:**
```bash
curl -LO https://github.com/speier/tokenscout/releases/latest/download/tokenscout_*_Linux_x86_64.tar.gz
tar -xzf tokenscout_*.tar.gz
./tokenscout version
```

**Windows:**
Download `.zip` from releases, extract, run `tokenscout.exe version` in PowerShell.

### Option 2: From Source
```bash
git clone https://github.com/speier/tokenscout
cd tokenscout
go build
./tokenscout version
```

## Configuration

### 1. Initialize
```bash
./tokenscout init
```
Creates `config.yaml` with defaults.

### 2. Configure RPC (Required for Production)

**Helius (Recommended - $29/month):**

1. Sign up at https://helius.dev
2. Create API key from dashboard
3. Edit `config.yaml`:

```yaml
solana:
  rpc_url: https://mainnet.helius-rpc.com/?api-key=YOUR_API_KEY
  ws_url: wss://mainnet.helius-rpc.com/?api-key=YOUR_API_KEY
```

**Why Helius:**
- No rate limits on paid plans
- Enhanced APIs (parsed transactions, webhooks)
- Best performance for trading bots
- Developer plan: 1M requests/day

**Free RPC (Testing Only):**
```yaml
solana:
  rpc_url: https://api.mainnet-beta.solana.com
  ws_url: wss://api.mainnet-beta.solana.com
```
⚠️ Rate limited (~40 req/10s) - will miss most tokens

### 3. Adjust Trading Rules
```yaml
rules:
  min_holders: 10              # Minimum token holders
  max_top_holder_pct: 30       # Max % one wallet can hold
  max_mint_age_sec: 300        # Only tokens < 5min old
  require_no_freeze: true      # Reject if can freeze
  require_no_mint: true        # Reject if can mint more

risk:
  stop_loss_pct: -10           # Exit at -10% loss
  take_profit_pct: 10          # Exit at +10% profit
  max_trade_duration_sec: 600  # Exit after 10 minutes
```

## Wallet Setup

### Create New Wallet
```bash
./tokenscout wallet new
```
Saves to `wallet.json` (keep this safe!).

### Import Existing Wallet
Place your Solana wallet JSON in `wallet.json`, or configure path:
```yaml
solana:
  wallet_path: /path/to/your/wallet.json
```

### Fund Wallet
Send SOL to your wallet address:
```bash
./tokenscout wallet show
# Displays: Wallet Address: ABC123...XYZ
```

## Running

### Dry-Run (Simulation Only)
Safe way to test - no real trades:
```bash
./tokenscout start --dry-run
```

Shows what would happen with real quotes from Jupiter.

### Live Trading
⚠️ **Uses real SOL:**
```bash
./tokenscout start
```

### View Status
```bash
./tokenscout status      # Statistics
./tokenscout positions   # Open positions
./tokenscout trades      # Trade history
```

### Emergency Stop
Close all positions immediately:
```bash
./tokenscout sellall
```

## Monitoring

### Logs
- **INFO level** (default): Clean, user-friendly messages
- **DEBUG level**: Technical details including file:line

```bash
./tokenscout start --log-level debug
```

### Database
All data stored in `tokenscout.db` (SQLite):
- Trades
- Positions
- Events
- Blacklist/Whitelist

## Troubleshooting

**No tokens detected:**
- **Most common:** Using free RPC with rate limits (~40 req/10s)
- **Solution:** Use Helius or another paid RPC service
- Free RPC will miss 95%+ of tokens due to rate limiting

**Rate limit errors (429):**
- Fetching every transaction hits limits quickly
- **Solution:** Switch to Helius ($29/month) for unlimited requests
- Alternative: Wait for off-peak hours (late night UTC)

**WebSocket disconnects:**
- Normal behavior, bot auto-reconnects
- Check logs for connection status

**API Key not working:**
- Verify key is correct in `config.yaml`
- Check Helius dashboard for usage limits
- Make sure format is: `https://mainnet.helius-rpc.com/?api-key=YOUR_KEY`

## Configuration Reference

Full `config.yaml` structure:

```yaml
engine:
  mode: dry_run              # dry_run or live

solana:
  rpc_url: https://api.mainnet-beta.solana.com
  ws_url: wss://api.mainnet-beta.solana.com
  wallet_path: wallet.json
  jupiter_api_url: https://quote-api.jup.ag/v6

listener:
  enabled: true
  mode: websocket            # websocket or polling
  polling_interval_sec: 10   # For polling mode
  programs:                  # DEXes to monitor
    - "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"  # Raydium
    - "9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"  # Orca

trading:
  max_spend_per_trade: 0.5   # SOL per trade
  max_open_positions: 3
  slippage_bps: 150          # 1.5%

rules:
  min_holders: 10
  max_top_holder_pct: 30
  max_mint_age_sec: 300
  require_no_freeze: true
  require_no_mint: true

risk:
  stop_loss_pct: -10
  take_profit_pct: 10
  max_trade_duration_sec: 600
```
