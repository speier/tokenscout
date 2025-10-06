# Agent Instructions

Guidelines for AI agents working on this codebase.

## Architecture Overview

This is a **Solana trading bot** with a pipeline architecture for monitoring DEX pools and executing automated trades:

```
Listener → Processor → Executor
           ↓           ↓
        RuleEngine   Monitor (positions)
```

**Core Components** (all in `internal/engine/`):
- **Listener/Poller/WebhookListener**: Ingest blockchain events from Raydium/Orca programs
- **Processor**: Deduplicates events (5-min window), evaluates rules, triggers buy decisions
- **Executor**: Executes buys/sells via Jupiter aggregator
- **Monitor**: Watches positions every 5s, enforces stop-loss/take-profit/time limits
- **RuleEngine**: Filters tokens based on liquidity, holders, authorities, age

**Data Flow**:
1. Listener detects new token pool on Raydium/Orca → sends to event channel
2. Processor deduplicates events → evaluates via RuleEngine
3. RuleEngine fetches metadata (holders, liquidity, mint age) from RPC
4. If rules pass → Executor gets Jupiter quote → creates trade/position in DB
5. Monitor polls positions every 5s → checks exits → triggers sells

**Key Files**:
- `internal/engine/engine.go` - Orchestration, component initialization
- `internal/engine/processor.go` - Event dedup, rule evaluation flow
- `internal/engine/executor.go` - Buy/sell logic, Jupiter integration
- `internal/strategies/presets.go` - Strategy config examples

## Strategy System

**Critical Pattern**: Strategies override config at runtime, NOT stored in YAML.

```go
// Flow: CLI flag → Config.Strategy → Executor → Database
./tokenscout start --strategy snipe_flip
```

- Every trade/position MUST have `strategy` field
- Sell trades inherit strategy from the position being closed
- Default to `"custom"` when no `--strategy` flag provided
- 5 built-in presets in `internal/strategies/presets.go`

## Database Patterns

**SQLite with idempotent migrations** (`internal/repository/sqlite.go`):

```go
// ALWAYS check column exists before adding
var hasColumn int
db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('table') WHERE name='column'`).Scan(&hasColumn)
if hasColumn == 0 {
    db.Exec(`ALTER TABLE table ADD COLUMN column TEXT DEFAULT '';`)
}
```

**Tables**: trades, positions, events, configs, blacklist, whitelist

**Key Conventions**:
- All trades/positions have `strategy` field (default `''` for backward compat)
- Timestamps stored as Unix epochs, converted to `time.Time` in models
- Use `COALESCE(strategy, '')` in queries for null safety

## Critical Conventions

### Listener Modes (3 options)
- **WebSocket**: Real-time, requires paid RPC (Helius)
- **Polling**: Polls RPC every N seconds, works with free RPC
- **Webhook**: Helius webhook endpoint (port 8080)

Set via `config.yaml`:
```yaml
listener:
  mode: websocket  # or "polling" or "webhook"
```

### Dry-Run Mode
- Engine mode: `dry_run` vs `live`
- Dry-run gets REAL Jupiter quotes but doesn't execute swaps
- Trade status: `"DRY_RUN"` in tx_sig field

### Error Handling
- No panics in production code
- Clear errors: `fmt.Errorf("failed to X: %w", err)`
- Log context: `logger.Error().Err(err).Str("mint", mint).Msg(...)`

### Jupiter Integration
- Always get real quote via `jupiterClient.GetQuote()` (even in dry-run)
- Quote has exact amounts, price impact, routes
- Calculate USD: `(inAmount / 1e9 * solPrice) / (outAmount / 1e9)`

## Documentation Standards

### README.md
- Keep stable and general (project overview only)
- No detailed command lists
- No references to other markdown files
- Max ~30-40 lines
- Focus: Features, quick start, license

### docs/USER_GUIDE.md
- Complete user documentation in one file
- Be concise and actionable
- Include: setup, commands, config, troubleshooting
- Target: Regular users, not developers

### Temporary Files
- Never commit: `*_SUMMARY.md`, `*_FIX.md`, implementation notes
- Clean up after feature completion

## Build & Test Commands

```bash
# Development: Use go run for ad-hoc runs (preferred)
go run . strategies compare
go run . start --dry-run --strategy snipe_flip
go run . positions

# Build binary (for releases/production only)
go build -o tokenscout .

# Test (no data needed)
go run . strategies compare

# List strategies
go run . start --list-strategies
```

## Best Practices

1. **Fish Shell**: Use printf/echo instead of heredocs
2. **Concise**: Keep code and docs brief but complete
3. **User-First**: Think from user perspective, not developer
4. **Test**: Always verify builds and migrations work
5. **Cleanup**: Remove temporary files after completion
6. **No Root Clutter**: No random .md files in root (use docs/ if needed)
