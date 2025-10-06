# Dry-Run Mode - Realistic Simulations

## Overview

TokenScout's dry-run mode now uses **real Jupiter quotes** to simulate trades with actual market conditions. This lets you test strategies and profitability before risking real funds.

## What's Real in Dry-Run?

### ✅ Real Data
- **Jupiter Quotes**: Actual quotes from Jupiter aggregator
- **Token Prices**: Real prices based on current liquidity pools
- **Slippage**: Shows actual price impact of your trade size
- **SOL/USD Price**: Live Solana price
- **PnL Calculations**: Based on real entry and exit prices

### ❌ What's Simulated
- **Trade Execution**: No actual swap transactions
- **Wallet Balance**: Doesn't deduct SOL from wallet
- **Position Quantities**: Simulated token holdings
- **Transaction Fees**: Not tracked

## How It Works

### Buy Flow
```
1. Detects new token on Raydium/Orca
2. Evaluates against rules
3. ✅ Approved → Fetch Jupiter quote
   "Buy 0.05 SOL worth of token XYZ"
   
4. Jupiter returns:
   - In: 0.05 SOL ($5.00)
   - Out: 1,234,567 tokens
   - Price Impact: 0.5%
   - Token Price: $0.00000405 per token
   
5. Records position with REAL price
6. NO actual transaction sent
```

### Sell Flow
```
1. Stop-loss/take-profit/time-limit triggered
2. Fetch Jupiter sell quote
   "Sell 1,234,567 tokens for SOL"
   
3. Jupiter returns:
   - In: 1,234,567 tokens  
   - Out: 0.06 SOL ($6.00)
   - Price Impact: 0.4%
   
4. Calculates profit: +$1.00 (+20%)
5. Closes position
6. NO actual transaction sent
```

## Example Output

```bash
INFO Real quote received from Jupiter 
  mint=ABC1..DEF2 
  sol_spent=0.05 
  tokens_received=1234567.000000000 
  token_price_usd=0.00000405 
  price_impact=0.005

INFO DRY RUN: Position opened with real quote price 
  mint=ABC1..DEF2 
  token_price_usd=0.00000405 
  tokens=1234567.000000000

... 10 minutes later ...

INFO Real sell quote received from Jupiter 
  mint=ABC1..DEF2 
  sol_received=0.06 
  usd_received=6.00 
  price_impact=0.004

INFO DRY RUN: Position closed with real sell price 
  mint=ABC1..DEF2 
  usd_received=6.00
```

## Benefits

### 1. Test Strategies Risk-Free
```bash
# Try aggressive settings
./tokenscout start --dry-run

# Edit config.yaml:
rules:
  min_holders: 5              # Lower threshold
  dev_wallet_max_pct: 30      # Higher risk tolerance

risk:
  stop_loss_pct: 5            # Tighter stop
  take_profit_pct: 50         # Higher target
```

See how often you'd hit stop-loss vs take-profit.

### 2. Measure Real Slippage
Large trades on low-liquidity tokens show high price impact:
```
Small trade (0.01 SOL): 0.1% impact
Large trade (1.0 SOL):  5.0% impact ← Would lose 5% immediately
```

### 3. Calculate Realistic Returns
```bash
# Check your simulated performance
./tokenscout status

# View all trades with real prices
./tokenscout trades --limit 50

# See positions with real entry prices
./tokenscout positions
```

### 4. Find Optimal Settings
Run for 24 hours with different configs:

**Config A (Conservative):**
```yaml
trading:
  max_spend_per_trade: 0.05
rules:
  min_holders: 100
risk:
  stop_loss_pct: 10
  take_profit_pct: 20
```

**Config B (Aggressive):**
```yaml
trading:
  max_spend_per_trade: 0.1
rules:
  min_holders: 20
risk:
  stop_loss_pct: 15
  take_profit_pct: 50
```

Compare results to find what works.

## Limitations

### Quote vs Execution Price
**Dry-run shows**: Quote price at that moment  
**Live trading gets**: Actual execution price (can differ due to MEV, frontrunning)

Real execution might be worse than quotes suggest.

### No Transaction Fees
Dry-run doesn't account for:
- Solana transaction fees (~0.000005 SOL)
- Jupiter platform fees
- Priority fees during congestion

Real trading has ~$0.01-0.10 per swap.

### Liquidity Can Change
Token quoted at $0.001 might be $0.0005 when you actually try to sell (rug pull, liquidity removed).

### Rate Limiting
Free RPC endpoints may block quote requests:
```
ERROR Failed to get Jupiter quote: rate limited
```

Solution: Use paid RPC or lower quote frequency.

## Best Practices

### 1. Run for 24-48 Hours
```bash
# Start dry-run and let it run
./tokenscout start --dry-run

# Check results next day
./tokenscout status
./tokenscout trades | grep "SELL"
```

See how many tokens pass filters and what returns look like.

### 2. Monitor Quote Failures
```bash
# Watch logs
./tokenscout start --dry-run --log-level info | grep "quote"
```

If many quote failures → RPC issues or token problems.

### 3. Test Edge Cases
```yaml
# Try extreme settings
trading:
  max_spend_per_trade: 1.0    # Large trades

rules:
  min_holders: 1              # Fresh tokens
```

See how price impact affects large orders.

### 4. Compare Multiple Runs
```bash
# Run 1: Conservative (day 1)
cp tokenscout.db results-conservative.db

# Run 2: Aggressive (day 2) 
rm tokenscout.db
./tokenscout init
# Edit config...
./tokenscout start --dry-run

# Compare databases
sqlite3 results-conservative.db "SELECT COUNT(*) FROM trades"
sqlite3 tokenscout.db "SELECT COUNT(*) FROM trades"
```

## Going Live

When dry-run looks profitable:

1. **Review all trades**
   ```bash
   sqlite3 tokenscout.db "SELECT * FROM trades WHERE status='EXECUTED'"
   ```

2. **Check success rate**
   - How many hit take-profit?
   - How many hit stop-loss?
   - What's the win/loss ratio?

3. **Start small**
   ```yaml
   trading:
     max_spend_per_trade: 0.01  # Start with $1 trades
   ```

4. **Remove --dry-run flag**
   ```bash
   ./tokenscout start  # LIVE MODE - uses real money!
   ```

5. **Monitor closely**
   Watch first few trades carefully.

## Summary

Dry-run mode is now **realistic** because it:
- ✅ Uses real Jupiter quotes
- ✅ Shows actual prices and slippage
- ✅ Calculates real PnL
- ✅ Tests strategies risk-free

But remember:
- ⚠️ Live execution can differ from quotes
- ⚠️ Doesn't include transaction fees
- ⚠️ Markets can change between quote and execution

**Use dry-run to validate strategies, then start live with small amounts!**
