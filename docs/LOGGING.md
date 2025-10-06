# Logging Guide

## Log Levels

### Info (Default)
Clean, actionable events only:
```
✓ WebSocket connected
→ Evaluating token: ABC1..DEF2
✗ Rejected: holders 5 < 10
💰 Buying: ABC1..DEF2
✓ DRY RUN: Buy simulated
📈 Take-profit triggered: +12.5%
📉 Stop-loss triggered: -8.2%
⏱ Max duration exceeded, selling
💸 Selling: ABC1..DEF2
```

### Debug
Verbose details for troubleshooting:
```bash
go run . start --dry-run --log-level debug
```

## Log Output Examples

### Normal Operation (Info Level)
```
INFO Running in DRY RUN mode - no trades will be executed
INFO Trading engine started mode=dry_run max_positions=3
WARN Failed to load wallet, execution disabled (create with: tokenscout wallet new)
INFO Starting blockchain listener programs=2
INFO ✓ WebSocket connected, subscribed to programs programs=2
INFO → Evaluating token mint=ABC1..DEF2
INFO ✗ Rejected mint=ABC1..DEF2 reason="holders: 5 < 10"
INFO → Evaluating token mint=XYZ3..QRS4
INFO 💰 Buying mint=XYZ3..QRS4
INFO ✓ DRY RUN: Buy simulated mint=XYZ3..QRS4
INFO 📈 Take-profit triggered mint=XYZ3..QRS4 pnl=12.5
INFO 💸 Selling mint=XYZ3..QRS4 reason=take_profit
INFO ✓ DRY RUN: Sell simulated mint=XYZ3..QRS4
```

**Note:** "Failed to load wallet" warning is expected in dry-run until you create a wallet with `tokenscout wallet new`.

### Quiet Mode (Warnings/Errors Only)
```bash
go run . start --dry-run --log-level warn
```
Only shows problems.

## Emojis Legend

- ✓ Success / Completed
- ✗ Rejected / Failed
- → Action in progress
- 💰 Buying
- 💸 Selling
- 📈 Profit (take-profit triggered)
- 📉 Loss (stop-loss triggered)
- ⏱ Time limit reached

## Common Rejection Reasons

```
"holders: 5 < 10"              # Not enough holders
"top holder: 85.0% > 20.0%"    # Dev wallet too concentrated
"has freeze authority"         # Token can be frozen
"has mint authority"           # Unlimited supply
"too old: 9000s"              # Token too old
"failed to fetch holders"      # RPC rate limit hit
"failed to fetch token info"   # Can't query token data
```

## Tips

**Less logging:**
```bash
--log-level warn  # Only warnings/errors
```

**More logging:**
```bash
--log-level debug  # Everything
```

**Default:**
```bash
--log-level info  # Clean, actionable events
```
