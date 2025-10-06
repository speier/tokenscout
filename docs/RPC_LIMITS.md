# Solana RPC Rate Limits

## Free Public RPC (api.mainnet-beta.solana.com)

### WebSocket Limits
| Limit | Value |
|-------|-------|
| Bandwidth | 100 MB per 30 seconds per IP |
| Concurrent Connections | 40 per IP |
| Connection Rate | 40 new connections per 10 seconds per IP |
| Subscriptions | 50 active per connection |

### HTTP Limits
- Varies by endpoint
- Generally more permissive than WebSocket
- Polling mode uses HTTP

## When You'll Hit Limits

### WebSocket
- Monitoring 10+ high-activity programs
- Multiple bot instances on same IP
- High-frequency trading (100+ events/minute)

### Solutions
1. **Use polling mode** (slower but no limits)
2. **Get paid RPC** (Helius, QuickNode, Triton)
3. **Use multiple IPs** (not recommended)
4. **Reduce monitored programs** (fewer subscriptions)

## Monitoring Rate Limits

Bot will log errors if hitting limits:
```
ERR WebSocket connection failed, retrying
```

If you see frequent reconnections, you're likely hitting limits.

## Recommendations

### Development/Testing
- ✅ Use WebSocket with free RPC
- ✅ Monitor 2-3 programs (Raydium + Orca)
- ✅ Single bot instance

### Production
- ✅ Get paid RPC if high volume
- ✅ Or use polling mode (reliable but slower)
- ✅ Monitor connection errors

## Paid RPC Comparison

| Provider | Free Tier | Price | WebSocket |
|----------|-----------|-------|-----------|
| Helius | 100k credits | $49/mo | Unlimited |
| QuickNode | Limited | $9-49/mo | Unlimited |
| Triton | Trial | Custom | Unlimited |

All remove rate limits and provide better uptime.
