# Testing Guide

## Quick Start

```bash
# Build
go build

# Or run directly without building
go run . --help

# Initialize config
./tokenscout init

# Create wallet (optional, for live trading)
./tokenscout wallet new

# Start in dry-run mode (safe, simulates trades)
./tokenscout start --dry-run --log-level debug
```

## What Happens

1. Connects to Solana WebSocket
2. Monitors Raydium & Orca programs
3. Parses events in real-time
4. Checks each token against rules:
   - ❌ Blacklist
   - ❌ Freeze authority
   - ❌ Mint authority  
   - ✓ Min holders (10+)
   - ✓ Dev concentration (<20%)
5. Simulates buying if rules pass
6. Auto-sells after 10 minutes

## Check Status

```bash
# View stats
./tokenscout status

# List positions
./tokenscout positions

# List trades
./tokenscout trades --limit 20

# Emergency sell all
./tokenscout sellall

# Or use go run .
go run . status
go run . positions
```

## Configuration

Edit `config.yaml`:

```yaml
# Lower holder requirement for fresh tokens
rules:
  min_holders: 5  # Very low for testing

# Shorter exit time for testing
risk:
  max_trade_duration_sec: 120  # 2 minutes instead of 10
```

## Watching Logs

```bash
# Debug mode (verbose)
go run . start --dry-run --log-level debug

# Info mode (cleaner)
go run . start --dry-run --log-level info
```

## Expected Log Output

```
✓ WebSocket connected
✓ Subscribed to program logs (Raydium)
✓ Subscribed to program logs (Orca)
→ New event detected
→ Evaluating rules...
✗ Token rejected: insufficient holders
→ Token passed rules, executing buy
✓ DRY RUN: Position opened
→ Position exceeded max duration, selling
✓ DRY RUN: Position closed
```

## Testing Different Scenarios

### Disable listener (no trading)
```yaml
listener:
  enabled: false
```

### Test polling mode (slower but reliable)
```yaml
listener:
  mode: polling
  polling_interval_sec: 30
```

### Very aggressive (more trades)
```yaml
rules:
  min_holders: 5
  dev_wallet_max_pct: 50
  allow_mint_authority: true
```

## Troubleshooting

**No events detected:**
- Check WebSocket connected successfully
- Wait 1-2 minutes (Raydium/Orca aren't super active)
- Try different programs in config

**All tokens rejected:**
- Lower `min_holders` in config
- Increase `dev_wallet_max_pct`
- Check logs for rejection reasons

**Rate limit errors:**
- Switch to polling mode
- Or get paid RPC (Helius/QuickNode)
