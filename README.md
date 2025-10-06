# TokenScout

Solana trading bot that monitors new tokens and trades automatically based on configurable rules.

> **Quick Start:** [QUICKSTART.md](QUICKSTART.md) - Get running in 60 seconds  
> **Windows Users:** [Windows Quick Start](docs/WINDOWS_QUICKSTART.md)

## Installation

### From Release (Recommended)
Download the latest binary from [Releases](https://github.com/speier/tokenscout/releases).

```bash
# macOS / Linux
tar -xzf tokenscout_*.tar.gz
./tokenscout init

# Windows - Extract .zip file, then:
tokenscout.exe init
```

## Quick Start

```bash
# Run directly from source (no build needed)
go run . init

# Create wallet
go run . wallet new

# Start in dry-run mode (safe)
go run . start --dry-run

# Or build binary first
go build
./tokenscout start --dry-run
```

## Features

- 🔍 Monitor new tokens on Raydium/Orca
- 📊 Rule-based filtering (liquidity, holders, authorities)
- 💱 Jupiter DEX integration for swaps
- 🛡️ Risk management (stop-loss, take-profit, time limits)
- 📈 Position tracking in SQLite
- 🔄 Two listening modes: WebSocket (real-time) or Polling (fallback)

## Commands

```bash
tokenscout init                    # Create config.yaml
tokenscout start                   # Start trading (live mode)
tokenscout start --dry-run         # Test without executing trades
tokenscout status                  # Show stats
tokenscout trades --limit 20       # List recent trades
tokenscout positions               # Show open positions
tokenscout sellall                 # Emergency: close all positions
tokenscout wallet new              # Generate new wallet
tokenscout wallet show             # View address & balance
tokenscout version                 # Show version info
tokenscout --version               # Show version (short)

# Or use go run . for any command
go run . init
go run . start --dry-run
```

## Configuration

Edit `config.yaml`:

```yaml
listener:
  mode: polling  # "polling" (free RPC) or "websocket" (paid RPC)
  polling_interval_sec: 10

trading:
  max_spend_per_trade: 0.5  # SOL per trade
  slippage_bps: 150

risk:
  stop_loss_pct: 10
  take_profit_pct: 10
  max_trade_duration_sec: 600  # Exit after 10 min
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - System design and components
- [Configuration](docs/CONFIGURATION.md) - All config options explained
- [Testing](docs/TESTING.md) - How to test dry-run mode
- [Logging](docs/LOGGING.md) - Log levels and output examples
- [Makefile](docs/MAKEFILE.md) - All make commands reference
- [Version Info](docs/VERSION.md) - Version and build information
- [Quick Release](docs/QUICKSTART_RELEASE.md) - TL;DR: `make release VERSION=v1.0.0`
- [Releasing](docs/RELEASING.md) - Full release documentation
- [Development](docs/DEVELOPMENT.md) - Status and roadmap

## Safety

⚠️ **Always start with dry-run mode**  
⚠️ **Use small amounts** (0.1-0.5 SOL per trade)  
⚠️ **Test on devnet first**

## License

MIT
