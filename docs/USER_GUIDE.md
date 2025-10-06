# TokenScout User Guide

## Installation

**macOS/Linux:**
```bash
curl -LO https://github.com/speier/tokenscout/releases/latest/download/tokenscout_*_Darwin_arm64.tar.gz
tar -xzf tokenscout_*.tar.gz
./tokenscout version
```

**From source:**
```bash
git clone https://github.com/speier/tokenscout
cd tokenscout
go build
```

## Setup

### 1. Initialize Configuration
```bash
./tokenscout init
```

Creates two files:
- `config.yaml` - Trading rules and behavior settings
- `.env` - RPC URLs and API keys (keep this private!)

### 2. Configure RPC URLs

Edit `.env` and add your Helius API key:
```bash
HELIUS_API_KEY=your_actual_api_key_here
```

The key will be automatically injected into the RPC URLs defined in `config.yaml`.

**RPC Providers:**
- **Helius**: Best for bots, no rate limits on paid plans ($29/mo)
- **Public RPC**: Free but heavily rate limited (polling mode only)

For non-Helius providers, you can override the full URLs:
```bash
SOLANA_RPC_URL=https://your-rpc-provider.com
SOLANA_WS_URL=wss://your-rpc-provider.com
```

### 3. Create/Import Wallet
```bash
# Create new wallet
./tokenscout wallet new

# Or import existing wallet
cp /path/to/wallet.json wallet.json
```

## Usage

### Basic Commands

```bash
# Start in simulation mode (no real trades)
./tokenscout start --dry-run

# Start with a specific strategy
./tokenscout start --strategy snipe_flip --dry-run

# List available strategies
./tokenscout start --list-strategies

# View open positions
./tokenscout positions

# View trade history
./tokenscout trades

# Compare strategy performance
./tokenscout strategies compare

# Close all positions (emergency)
./tokenscout sellall

# View wallet balance
./tokenscout wallet show
```

### Strategies

TokenScout includes 5 built-in strategies:

| Strategy | Hold Time | Entry | Exit | Risk |
|----------|-----------|-------|------|------|
| `snipe_flip` | 3-5 min | Very early | +18%/-8% | High |
| `conservative` | 10-20 min | Established | +25%/-10% | Low |
| `scalping` | 30s-2min | Ultra early | +10%/-5% | Highest |
| `momentum_rider` | 5-15 min | Early volume | +40%/-15% | Medium |
| `data_collection` | Observe only | All tokens | Track only | Zero |

**Using preset strategies:**
```bash
./tokenscout start --strategy <name>
./tokenscout start --strategy snipe_flip --dry-run
```

**Using custom strategy configs:**
```bash
./tokenscout start --strategy-config strategies/fast_flip_conservative.yaml --dry-run
```

Create your own in `strategies/my_strategy.yaml`:
```yaml
rules:
  min_liquidity_usd: 5000
  max_mint_age_sec: 600
  
risk:
  take_profit_pct: 15
  stop_loss_pct: 7
```

Strategy configs only override specified values - everything else uses `config.yaml` defaults.

### Configuration

Key settings in `config.yaml`:

**Trading:**
```yaml
trading:
  max_spend_per_trade: 0.2    # SOL per trade
  max_open_positions: 5       # Concurrent positions
  slippage_bps: 400          # 4% slippage tolerance
```

**Risk Management:**
```yaml
risk:
  stop_loss_pct: 8           # Exit at -8% loss
  take_profit_pct: 18        # Exit at +18% profit
  max_trade_duration_sec: 240 # Max 4 minutes hold
```

**Token Filters:**
```yaml
rules:
  min_holders: 3             # Minimum holders
  min_liquidity_usd: 3000    # Minimum liquidity
  max_mint_age_sec: 300      # Only tokens < 5min old
  block_freeze_authority: true
  allow_mint_authority: false
```

## Going Live

1. Fund your wallet with SOL
2. Set proper RPC endpoints (Helius recommended)
3. Test with `--dry-run` first
4. Remove `--dry-run` flag to go live:
   ```bash
   ./tokenscout start --strategy snipe_flip
   ```

⚠️ **Warning**: Real money at risk. Start small, test thoroughly.

## Troubleshooting

**No tokens detected:**
- Check RPC endpoints are working
- Verify WebSocket connection
- Public RPCs are heavily rate limited

**Trades failing:**
- Increase `slippage_bps` for faster tokens
- Check wallet has enough SOL
- Verify Jupiter API is accessible

**Database errors:**
- Delete `tokenscout.db` to reset
- Check file permissions

## Data Storage

- `config.yaml` - Trading rules and behavior (safe to commit)
- `.env` - RPC URLs and API keys (**never commit this!**)
- `wallet.json` - Your wallet keypair (**keep this safe!**)
- `tokenscout.db` - SQLite database (trades, positions)

## Performance Tracking

View strategy performance:
```bash
./tokenscout strategies compare
```

Shows win rate, avg profit, total trades per strategy.

## Safety Tips

1. Always test with `--dry-run` first
2. Start with small `max_spend_per_trade`
3. Use `stop_loss_pct` to limit losses
4. Monitor with `./tokenscout status`
5. Keep wallet seed phrase safe
6. **Never commit `.env` or `wallet.json` to git**
7. Use `.gitignore` to protect sensitive files
