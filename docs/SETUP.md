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

### 2. Configure RPC (Optional)
Edit `config.yaml`:
```yaml
solana:
  rpc_url: https://api.mainnet-beta.solana.com  # Free (limited)
  # Or use paid RPC for better reliability:
  # rpc_url: https://your-quicknode-url.com
```

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
- Free RPC has rate limits (~40 req/10s)
- Consider paid RPC (Helius, QuickNode, Triton)
- Or wait for off-peak hours

**WebSocket disconnects:**
- Normal behavior, auto-reconnects
- Free RPC limits: 100MB/30s, 40 connections

**Rate limit errors (429):**
- Reduce polling frequency
- Switch to paid RPC
- Bot fetches every transaction to parse - very RPC intensive

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
