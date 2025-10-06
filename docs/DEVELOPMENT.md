# Development Guide

## Architecture

```
┌─────────────┐
│  WebSocket  │ ← Solana RPC (Raydium/Orca program logs)
└──────┬──────┘
       │
       ↓
┌─────────────┐
│  Listener   │ ← Fetches full transactions
└──────┬──────┘
       │
       ↓
┌─────────────┐
│  Parsers    │ ← Modular DEX-specific parsers
└──────┬──────┘   (Raydium, Orca, extensible)
       │
       ↓
┌─────────────┐
│   Rules     │ ← Filters tokens (holders, age, authorities)
└──────┬──────┘
       │
       ↓
┌─────────────┐
│  Executor   │ ← Jupiter swaps (dry-run or live)
└──────┬──────┘
       │
       ↓
┌─────────────┐
│  Monitor    │ ← Tracks positions, exits (stop-loss, take-profit)
└─────────────┘
```

## Project Structure

```
├── internal/
│   ├── cli/         # Cobra commands
│   ├── config/      # Viper config loading
│   ├── engine/      # Trading engine + parsers
│   ├── logger/      # Zerolog setup
│   ├── models/      # Data structures
│   ├── repository/  # SQLite + interface
│   └── solana/      # RPC, Jupiter, wallet
├── docs/            # Documentation
├── .github/         # CI/CD workflows
├── main.go          # Entry point
├── Makefile         # Build automation
└── VERSION          # Single source of truth
```

## Building

```bash
# Development
go run . start --dry-run

# Production build
make build          # Builds with version info

# Test
make test

# Release (auto-increments version)
make release        # Bumps patch, tags, pushes
```

## Adding a New DEX Parser

Create parser in `internal/engine/parsers.go`:

```go
type PumpFunParser struct{}

func (p *PumpFunParser) Name() string {
    return "Pump.fun"
}

func (p *PumpFunParser) CanParse(programID solana.PublicKey, accounts []solana.PublicKey, data []byte) bool {
    pumpFunProgram := solana.MustPublicKeyFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P")
    return programID.Equals(pumpFunProgram) && len(accounts) >= 8
}

func (p *PumpFunParser) ParseTokenMint(accounts []solana.PublicKey, data []byte) (string, bool) {
    // Extract mint from accounts based on Pump.fun's layout
    mint := accounts[3] // Example: adjust based on actual layout
    return mint.String(), true
}
```

Add to registry:
```go
func NewParsersRegistry() *ParsersRegistry {
    return &ParsersRegistry{
        parsers: []InstructionParser{
            &RaydiumParser{},
            &OrcaParser{},
            &PumpFunParser{},  // ← Add here
        },
    }
}
```

## Releasing

### Automatic (Recommended)
```bash
make release        # Bumps patch: 0.1.0 → 0.1.1
make bump-minor     # Bumps minor: 0.1.0 → 0.2.0
make bump-major     # Bumps major: 0.1.0 → 1.0.0
```

### Manual
```bash
echo "0.2.0" > VERSION
make release-manual
```

GitHub Actions builds binaries for:
- macOS (Intel + Apple Silicon)
- Linux (x86_64 + arm64)
- Windows (x86_64)

## Current Limitations

### Known Issues
1. **Rate Limiting**: Free Solana RPC gets rate-limited quickly
   - Solution: Implement request queue or use paid RPC
2. **Parser Accuracy**: Account indices may vary between DEX versions
   - Solution: Log discriminators, test with real transactions
3. **No Honeypot Detection**: Can't detect if token is sellable
   - Solution: Simulate sell before buying

### Roadmap
- [ ] Implement request queue for rate limiting
- [ ] Add honeypot detection (sell simulation)
- [ ] Add min_liquidity_usd check
- [ ] Live trading mode (currently dry-run only)
- [ ] Web dashboard (basic CLI only now)
- [ ] Historical backtesting
- [ ] Multiple wallet support

## Testing

```bash
# Unit tests
go test ./...

# Dry-run test (safe, uses real RPC)
./tokenscout start --dry-run --log-level debug

# Check version injection
./tokenscout version
```

## CI/CD

Three workflows:
1. **CI** - Runs on every push (builds + tests)
2. **Release** - Triggers on tag push (publishes binaries)
3. **Test Release** - Validates GoReleaser config on PRs

## Contributing

1. Fork the repo
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes with conventional commits:
   - `feat: add new parser`
   - `fix: handle rate limits`
   - `docs: update setup guide`
4. Push to branch
5. Open Pull Request

## Getting Help

- Open an issue on GitHub
- Check existing issues for similar problems
- Include logs with `--log-level debug`
