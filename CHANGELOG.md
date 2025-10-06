# Changelog

## Phase 3.1 - Price & Age Features (2025-10-06)

### Added
- **Jupiter Price API integration**
  - Real HTTP requests to Jupiter Price API
  - Returns USD prices directly
  - Fallback to quote-based estimation if API fails
  
- **SOL/USD price fetching**
  - Real-time SOL price from Jupiter
  - Used in PnL calculations and conversions
  - Fallback to $100 if API unavailable
  
- **Token age verification**
  - Fetches first transaction timestamp
  - Calculates age in seconds
  - Rule engine checks against max_mint_age_sec
  - Rejects tokens older than configured threshold

### Improved
- Price tracking more accurate with direct API calls
- PnL calculations use real-time SOL prices
- Better error handling with fallbacks

### Configuration
```yaml
rules:
  max_mint_age_sec: 7200  # Reject tokens older than 2 hours
```

---

## Phase 3 - Token Validation & Execution (2025-10-06)

### Added
- **Token validation engine**
  - Fetch token info (authorities, decimals, supply)
  - Check freeze authority (block if present)
  - Check mint authority (block if present)
  - Count token holders via RPC
  - Analyze holder distribution (top holder percentage)
  
- **Execution engine**
  - Buy order execution (dry-run mode working)
  - Sell order execution (dry-run mode working)
  - Position tracking in database
  - Trade record creation with status tracking
  
- **Position monitoring**
  - Automatic monitoring every 5 seconds
  - Time-based exits (force sell after max duration)
  - Closes positions exceeding max_trade_duration_sec
  
- **CLI commands**
  - `bot sellall` - Emergency command to close all positions
  
### Rule Engine Updates
- Integrated token info fetching
- Added freeze authority check
- Added mint authority check
- Added holder count verification
- Added dev wallet concentration check
- Improved decision logging with specific reasons

- **Price tracking & risk management**
  - PnL calculation from entry price to current price
  - Stop-loss trigger at -10% loss
  - Take-profit trigger at +10% gain
  - Price fetching via Jupiter quotes (temporary solution)
  - Helper functions for SOL/USD conversion

### Configuration Changes
- Lowered `min_holders` from 200 → 10 (catch fresh tokens)
- Default risk settings: 10% stop-loss, 10% take-profit, 10min max duration

### Pending (Live Trading)
- Actual Jupiter swap execution (dry-run only for now)
- Jupiter Price API HTTP integration (currently using quotes as fallback)
- DEX pool liquidity checking
- Token age verification
- Honeypot detection

---

## Phase 2 - Trading Logic (2025-10-06)

### Added
- **Blockchain event listening**
  - WebSocket mode (free RPC with rate limits)
  - HTTP polling mode (fallback)
  - Raydium & Orca program monitoring
  - Event storage in database
  - Auto-reconnect on failure
  
- **Jupiter DEX integration**
  - Quote API client
  - Swap transaction building
  - Transaction signing
  - Priority fee support
  
- **Rule engine framework**
  - Blacklist/whitelist checking
  - Decision tracking with reasons
  - Integration with event processor

### Configuration
- Added `solana` config section (RPC URLs, wallet path)
- Added `listener` config (mode, programs, polling interval)
- WebSocket as default (works with free RPC)

### Documentation
- Cleaned up docs into `/docs` folder
- Added RPC limits documentation
- Corrected WebSocket availability info

---

## Phase 1 - Foundation (2025-10-06)

### Added
- Go project structure
- SQLite repository with migrations
- YAML configuration (Viper)
- Cobra CLI framework
- Wallet management (create, import, show)
- Structured logging (zerolog)
- Core data models (Trade, Position, Event, Config)

### Commands
- `bot init` - Create config.yaml
- `bot start` - Start trading engine
- `bot status` - Show stats
- `bot trades` - List trades
- `bot positions` - Show positions
- `bot wallet new/show/import` - Wallet management

### Documentation
- README with quick start
- Architecture docs
- Configuration reference
- Development status tracking
