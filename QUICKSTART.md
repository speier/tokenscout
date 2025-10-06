# TokenScout - Quick Start

Get up and running in 60 seconds.

## 1. Download

Go to [Releases](https://github.com/speier/tokenscout/releases/latest) and download for your OS:

- **Windows**: `tokenscout_v0.1.0_Windows_x86_64.tar.gz`
- **macOS (Apple Silicon)**: `tokenscout_v0.1.0_Darwin_arm64.tar.gz`
- **macOS (Intel)**: `tokenscout_v0.1.0_Darwin_x86_64.tar.gz`
- **Linux**: `tokenscout_v0.1.0_Linux_x86_64.tar.gz`

## 2. Extract

**Windows**: Right-click → Extract All (or use 7-Zip)  
**macOS/Linux**: `tar -xzf tokenscout_*.tar.gz`

## 3. Run

Open terminal/PowerShell in the extracted folder:

```bash
# Create config
./tokenscout init

# Create wallet
./tokenscout wallet new

# Start monitoring (safe mode - no real trades)
./tokenscout start --dry-run
```

**Windows users:** Add `.exe` → `.\tokenscout.exe init`

## 4. Watch It Work

```
INFO ✓ WebSocket connected
INFO → Evaluating token mint=ABC1..DEF2
INFO ✗ Rejected reason="holders: 5 < 10"
```

Press `Ctrl+C` to stop.

## Next Steps

**Edit config.yaml** to customize:
- Rules (minimum holders, liquidity, etc.)
- Risk management (stop-loss, take-profit)
- Position limits

**View activity:**
```bash
./tokenscout status    # Stats
./tokenscout trades    # Recent evaluations
./tokenscout positions # Open positions (if any)
```

**Go live** (real trades):
```bash
# Fund your wallet first!
./tokenscout wallet show

# Then remove --dry-run flag
./tokenscout start
```

⚠️ **Warning:** Live mode executes real trades with real money. Test thoroughly in dry-run first!

## Platform-Specific Guides

- [Windows Detailed Guide](docs/WINDOWS_QUICKSTART.md)
- [Full Documentation](README.md)
- [Testing Guide](TEST_COMMANDS.md)
- [Configuration Reference](docs/CONFIGURATION.md)

## Minimum Commands (TL;DR)

```bash
./tokenscout init
./tokenscout wallet new
./tokenscout start --dry-run
```

Done! 🚀
