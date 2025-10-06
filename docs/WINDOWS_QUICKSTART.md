# Windows Quick Start

## Download

1. Go to: https://github.com/speier/tokenscout/releases/latest
2. Download: `tokenscout_v0.1.0_Windows_x86_64.tar.gz`
3. Extract it (right-click → Extract All, or use 7-Zip)

## Run

Open PowerShell in the extracted folder:

```powershell
# Create config file
.\tokenscout.exe init

# Create wallet
.\tokenscout.exe wallet new

# Start bot (safe mode - simulates trades)
.\tokenscout.exe start --dry-run
```

## What You'll See

```
INFO ✓ WebSocket connected, subscribed to programs
(waits for new tokens...)
INFO → Evaluating token mint=ABC1..DEF2
INFO ✗ Rejected reason="holders: 5 < 10"
```

Press `Ctrl+C` to stop.

## That's It!

The bot is now monitoring the Solana blockchain for new tokens and testing them against your rules (in dry-run mode, so no real trades happen).

## Edit Settings

Edit `config.yaml` to change:
- Minimum holders
- Stop loss / take profit percentages
- Maximum positions
- Time limits

## View Activity

```powershell
# Check status
.\tokenscout.exe status

# View evaluated tokens
.\tokenscout.exe trades

# Check version
.\tokenscout.exe version
```

## Need Help?

See full documentation:
- [Configuration Guide](CONFIGURATION.md)
- [Testing Guide](../TEST_COMMANDS.md)
- [Architecture](ARCHITECTURE.md)

## Common Issues

**"Windows protected your PC"**
- Click "More info" → "Run anyway"
- This happens because the binary isn't code-signed

**Can't extract .tar.gz?**
- Use 7-Zip: https://www.7-zip.org/
- Or WinRAR: https://www.win-rar.com/

**PowerShell says "running scripts is disabled"**
- Use Command Prompt (cmd) instead
- Or run: `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`
