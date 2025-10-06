# TokenScout

Solana trading bot that monitors new tokens on Raydium/Orca DEXes and executes trades automatically.

## Quick Start

```bash
# Download and setup
curl -LO https://github.com/speier/tokenscout/releases/latest/download/tokenscout_*_Darwin_arm64.tar.gz
tar -xzf tokenscout_*.tar.gz

# Initialize config files
./tokenscout init

# Add your RPC URLs to .env
# Edit .env and set SOLANA_RPC_URL and SOLANA_WS_URL

# Create wallet
./tokenscout wallet new

# Start trading (simulation mode)
./tokenscout start --dry-run
```

## Features

- 🔍 Real-time token monitoring (Raydium/Orca)
- 💰 Automated trading via Jupiter DEX
- 🛡️ Risk management (stop-loss, take-profit, time limits)
- 📊 Token filtering (liquidity, holders, authorities, age)
- 🎯 5 built-in trading strategies
- 📈 Performance tracking and analytics

## Documentation

See `docs/USER_GUIDE.md` for setup and usage instructions.

## License

MIT
