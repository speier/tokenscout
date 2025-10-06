# Architecture

## Overview

```
┌─────────────────────────┐
│  CLI (Cobra)            │
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│  Engine                 │
│  ├─ Listener/Poller     │ Monitor blockchain
│  ├─ Processor           │ Event pipeline
│  ├─ Rules               │ Filter tokens
│  └─ Executor (TODO)     │ Execute trades
└───────────┬─────────────┘
            │
┌───────────▼─────────────┐
│  Repository (SQLite)    │ Trades, positions, events
└─────────────────────────┘
```

## Components

### Listener/Poller
- **WebSocket**: Real-time events (free with rate limits)
  - 100 MB/30s bandwidth per IP
  - 40 concurrent connections per IP
  - 50 subscriptions per connection
- **Polling**: HTTP polling every 10s (fallback if hitting limits)
- Monitors Raydium and Orca programs
- Stores events in database

### Processor
- Consumes events from listener
- Deduplicates within 200ms window
- Sends to rule engine
- Triggers execution for approved tokens

### Rule Engine
- Evaluates tokens against filters
- Checks blacklist/whitelist
- Returns decision with reasons
- TODO: Liquidity, holders, authorities, honeypot detection

### Executor (TODO)
- Execute buy/sell via Jupiter
- Track positions
- Monitor exit conditions (stop-loss, take-profit, time)

### Repository
- SQLite for local storage
- Tables: trades, positions, events, blacklist, whitelist
- Swappable to Postgres for production

## Event Flow

```
Blockchain → Listener → Event Channel → Processor → Rules → Executor → Database
```

## Configuration

Two modes:
1. **WebSocket** (default): Free with rate limits, <1s latency
2. **Polling** (fallback): No rate limits, 10s latency

Free Solana RPC supports both modes. Paid RPC removes rate limits.
