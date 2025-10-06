# TokenScout

Solana trading bot that monitors new tokens on Raydium/Orca DEXes and trades automatically with configurable rules.

## Quick Start

```bash
# Install from release
curl -LO https://github.com/speier/tokenscout/releases/latest/download/tokenscout_*_Darwin_arm64.tar.gz
tar -xzf tokenscout_*.tar.gz

# Setup
./tokenscout init          # Creates config.yaml
./tokenscout wallet new    # Creates wallet

# Run (dry-run = simulation only, no real trades)
./tokenscout start --dry-run
```

**Or run from source:**
```bash
go run . init
go run . wallet new
go run . start --dry-run
```

## Features

- 🔍 Monitors Raydium/Orca for new token pools
- 💰 Auto-trades with Jupiter DEX integration
- 🛡️ Risk management (stop-loss, take-profit, time limits)
- 📊 Rule-based filtering (holders, liquidity, authorities, age)
- 📈 SQLite position/trade tracking
- 🔄 WebSocket + HTTP polling support

## Documentation

- [Setup Guide](docs/SETUP.md) - Installation, configuration, usage
- [Development](docs/DEVELOPMENT.md) - Contributing, architecture, roadmap

## Commands

```bash
init              Create default config
wallet new        Generate new wallet
wallet show       Display wallet address
start             Start trading (add --dry-run for simulation)
status            Show trading statistics
positions         List open positions
trades            Show trade history
sellall           Emergency: close all positions
version           Show version info
```

## License

MIT - See [LICENSE](LICENSE)
