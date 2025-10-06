# Configuration

## config.yaml

### Engine
```yaml
engine:
  mode: dry_run  # dry_run or live
  max_positions: 3
```

### Solana
```yaml
solana:
  rpc_url: https://api.mainnet-beta.solana.com  # Free RPC
  ws_url: wss://api.mainnet-beta.solana.com     # For WebSocket mode
  wallet_path: wallet.json
  jupiter_api_url: https://quote-api.jup.ag/v6
```

For paid RPC (Helius/QuickNode):
```yaml
solana:
  rpc_url: https://your-provider.com/YOUR_KEY
  ws_url: wss://your-provider.com/YOUR_KEY
```

### Listener
```yaml
listener:
  enabled: true
  mode: polling  # "polling" or "websocket"
  polling_interval_sec: 10  # Only for polling mode
  # DEX programs to monitor (these catch 90%+ of new tokens)
  programs:
    - "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"  # Raydium AMM (largest DEX)
    - "9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP"  # Orca Whirlpool (2nd largest)
  coalesce_window_ms: 200
```

**Programs:** These are Solana DEX smart contract addresses. The bot monitors their logs for new token pool creation events. You can add more DEXes, but Raydium + Orca cover the vast majority of new launches.

### Trading
```yaml
trading:
  base_mint: SOL
  quote_mint: USDC
  max_spend_per_trade: 0.5  # SOL amount per trade
  max_open_positions: 3
  slippage_bps: 150  # 1.5%
  priority_fee_microlamports: 5000
```

### Rules
```yaml
rules:
  min_liquidity_usd: 20000  # Minimum pool liquidity
  max_mint_age_sec: 7200    # Max 2 hours old
  min_holders: 10           # Lowered to catch fresh tokens
  dev_wallet_max_pct: 20    # Max % in one wallet
  block_freeze_authority: true
  allow_mint_authority: false
```

**Note:** `min_holders` was lowered from 200 to 10 to catch tokens early after launch.

### Risk
```yaml
risk:
  stop_loss_pct: 10         # Exit on 10% loss
  take_profit_pct: 10       # Exit on 10% profit
  max_trade_duration_sec: 600  # Force exit after 10 min
```

## Listening Modes

### WebSocket (Default - Free with Limits)
- Real-time (<1s latency)
- Works on free public RPC
- **Rate Limits:**
  - 100 MB per 30 seconds
  - 40 concurrent connections per IP
  - 50 subscriptions per connection
- Good for development and light usage

```yaml
listener:
  mode: websocket
```

### Polling (Fallback)
- Works with any RPC
- 10-second delay
- No rate limit concerns
- Use if hitting WebSocket limits

```yaml
listener:
  mode: polling
  polling_interval_sec: 10
```

## RPC Providers

### Free
- **Solana Public RPC**: Both HTTP and WebSocket available
  - `https://api.mainnet-beta.solana.com`
  - `wss://api.mainnet-beta.solana.com`
  - Rate limited (see above)

### Paid (No Rate Limits)
- **Helius**: helius.dev (100k free credits, then paid)
- **QuickNode**: quicknode.com (free tier available)
- **Triton**: triton.one
