# Quick Test Commands

## 1. Basic Functionality Test

```bash
# Build
make build

# Check version works
./tokenscout version

# Initialize config
./tokenscout init

# Check status
./tokenscout status
```

## 2. Start Bot in Dry-Run Mode

```bash
# Start monitoring (safe - simulates trades)
./tokenscout start --dry-run --log-level info

# Or with debug logs (verbose)
./tokenscout start --dry-run --log-level debug
```

**What you'll see:**
```
INFO Running in DRY RUN mode - no trades will be executed
INFO Trading engine started mode=dry_run max_positions=3
WARN Failed to load wallet, execution disabled
INFO Starting blockchain listener programs=2
INFO ✓ WebSocket connected, subscribed to programs programs=2

(waits for new tokens on Raydium/Orca...)

INFO → Evaluating token mint=ABC1..DEF2
INFO ✗ Rejected mint=ABC1..DEF2 reason="holders: 5 < 10"
```

## 3. Test in Another Terminal

While bot is running, open another terminal:

```bash
# Check status
./tokenscout status

# View positions (should be empty initially)
./tokenscout positions

# View trades (should be empty initially)
./tokenscout trades
```

## 4. Lower Thresholds to See More Activity

Edit `config.yaml`:

```yaml
rules:
  min_holders: 1              # Very low
  dev_wallet_max_pct: 99      # Almost anything
  allow_mint_authority: true  # Allow all
  block_freeze_authority: false  # Don't block
```

Then restart:
```bash
./tokenscout start --dry-run --log-level info
```

## 5. Test Without WebSocket (Polling Mode)

Edit `config.yaml`:

```yaml
listener:
  mode: polling
  polling_interval_sec: 30  # Check every 30 seconds
```

Restart:
```bash
./tokenscout start --dry-run
```

## 6. Expected Behavior

### On Free RPC (Expected Issues)
- ❌ Many "failed to fetch holders" rejections (rate limiting)
- ✅ WebSocket connects successfully
- ✅ Listens for events
- ⚠️ May not catch all tokens due to rate limits

### On Paid RPC (Better)
- ✅ Fetches holder data reliably
- ✅ Catches most new tokens
- ✅ Rules evaluated properly

## 7. Test Database

```bash
# View database
sqlite3 tokenscout.db

# In sqlite:
.tables
SELECT * FROM trades;
SELECT * FROM positions;
SELECT * FROM events;
.exit
```

## 8. Simulated Token Flow

When a new token is detected:
```
1. → Evaluating token mint=ABC1..DEF2
2. Fetching token info...
3. Checking freeze authority... ✓
4. Checking mint authority... ✓
5. Fetching holders... (may fail on free RPC)
6. Either:
   ✗ Rejected: reason
   OR
   💰 Buying
   ✓ DRY RUN: Buy simulated
```

## 9. Kill Test

```bash
# Ctrl+C to stop
^C
Shutting down gracefully...
INFO Trading engine shutting down
```

## 10. Clean Up Test Data

```bash
# Remove database
rm tokenscout.db

# Remove config (to reset)
rm config.yaml

# Start fresh
./tokenscout init
```

## Realistic Test Expectations

### Free RPC (api.mainnet-beta.solana.com)
- **Event detection**: ✅ Works
- **Token validation**: ⚠️ Often fails due to rate limits
- **Result**: You'll see many tokens rejected with "failed to fetch holders"

### Paid RPC (Helius/QuickNode)
- **Event detection**: ✅ Works great
- **Token validation**: ✅ Works reliably
- **Result**: Full rule evaluation, catches legitimate opportunities

## Quick "Did It Work?" Checklist

✅ Binary builds: `make build`  
✅ Version shows: `./tokenscout version`  
✅ Config creates: `./tokenscout init`  
✅ Status works: `./tokenscout status`  
✅ Bot starts: `./tokenscout start --dry-run`  
✅ WebSocket connects: See "✓ WebSocket connected"  
✅ Events detected: See "→ Evaluating token" (may take time)  
✅ Rules evaluated: See "✗ Rejected" or "💰 Buying"  
✅ Graceful shutdown: Ctrl+C works  

## Troubleshooting

**No tokens detected after 5 minutes?**
- Raydium/Orca aren't super active all the time
- Try during peak hours (US/EU trading times)
- Check logs for WebSocket errors

**All tokens rejected?**
```yaml
# Edit config.yaml - make it VERY permissive
rules:
  min_holders: 0
  dev_wallet_max_pct: 100
  allow_mint_authority: true
  block_freeze_authority: false
```

**Rate limit errors?**
- Expected on free RPC
- Switch to polling mode (slower but more reliable)
- Or upgrade to paid RPC
