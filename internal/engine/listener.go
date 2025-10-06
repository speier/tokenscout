package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
)

type Listener struct {
	wsURL    string
	programs []solana.PublicKey
	repo     repository.Repository
	eventCh  chan *models.Event
}

func NewListener(wsURL string, programIDs []string, repo repository.Repository) (*Listener, error) {
	programs := make([]solana.PublicKey, 0, len(programIDs))
	for _, id := range programIDs {
		pubkey, err := solana.PublicKeyFromBase58(id)
		if err != nil {
			return nil, fmt.Errorf("invalid program ID %s: %w", id, err)
		}
		programs = append(programs, pubkey)
	}

	return &Listener{
		wsURL:    wsURL,
		programs: programs,
		repo:     repo,
		eventCh:  make(chan *models.Event, 100),
	}, nil
}

func (l *Listener) Start(ctx context.Context) error {
	logger.Info().
		Str("ws_url", l.wsURL).
		Int("programs", len(l.programs)).
		Msg("Starting blockchain listener")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Listener shutting down")
			return nil
		default:
			if err := l.connect(ctx); err != nil {
				logger.Error().Err(err).Msg("Listener connection failed, retrying in 5s")
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}
}

func (l *Listener) connect(ctx context.Context) error {
	client, err := ws.Connect(ctx, l.wsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer client.Close()

	logger.Info().
		Int("programs", len(l.programs)).
		Msg("✓ WebSocket connected, subscribed to programs")

	// Subscribe to logs for each program
	for _, program := range l.programs {
		sub, err := client.LogsSubscribeMentions(
			program,
			rpc.CommitmentFinalized,
		)
		if err != nil {
			return fmt.Errorf("failed to subscribe to logs for %s: %w", program, err)
		}

		go l.handleSubscription(ctx, sub, program)
	}

	// Keep connection alive
	<-ctx.Done()
	return nil
}

func (l *Listener) handleSubscription(ctx context.Context, sub *ws.LogSubscription, program solana.PublicKey) {
	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			return
		default:
			got, err := sub.Recv(ctx)
			if err != nil {
				// Ignore context canceled (normal on shutdown)
				if ctx.Err() != nil {
					return
				}
				logger.Error().
					Err(err).
					Str("program", program.String()).
					Msg("Error receiving log")
				return
			}

			if got == nil {
				continue
			}

			l.processLog(ctx, got, program)
		}
	}
}

func (l *Listener) processLog(ctx context.Context, logResult *ws.LogResult, program solana.PublicKey) {
	if logResult.Value.Err != nil {
		// Skip failed transactions
		return
	}

	// Parse logs to detect new pool/token events
	event := l.parseEvent(logResult, program)
	if event == nil {
		return
	}

	// Store event in database
	if err := l.repo.CreateEvent(ctx, event); err != nil {
		logger.Error().
			Err(err).
			Str("mint", event.Mint).
			Msg("Failed to store event")
		return
	}

	logger.Info().
		Str("type", string(event.Type)).
		Str("mint", event.Mint).
		Str("program", program.String()).
		Msg("New event detected")

	// Send to event channel for processing
	select {
	case l.eventCh <- event:
	default:
		logger.Warn().Msg("Event channel full, dropping event")
	}
}

func (l *Listener) parseEvent(logResult *ws.LogResult, program solana.PublicKey) *models.Event {
	// Simple heuristic-based parsing
	// TODO: Implement proper instruction parsing for Raydium/Orca
	
	logs := logResult.Value.Logs
	signature := logResult.Value.Signature.String()

	// Look for keywords in logs
	for _, log := range logs {
		// Detect new pool initialization
		if containsAny(log, []string{"InitializePool", "initialize", "CreatePool"}) {
			return &models.Event{
				Type:      models.EventTypeNewPool,
				Mint:      extractMintFromLogs(logs),
				Timestamp: time.Now(),
				Raw:       toJSON(logResult),
			}
		}

		// Detect new token mint
		if containsAny(log, []string{"InitializeMint", "CreateMint"}) {
			return &models.Event{
				Type:      models.EventTypeNewMint,
				Mint:      extractMintFromLogs(logs),
				Timestamp: time.Now(),
				Raw:       toJSON(logResult),
			}
		}
	}

	// Log for debugging if we can't parse
	logger.Debug().
		Str("signature", signature).
		Str("program", program.String()).
		Strs("logs", logs).
		Msg("Unable to parse event from logs")

	return nil
}

func (l *Listener) EventChannel() <-chan *models.Event {
	return l.eventCh
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) && contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractMintFromLogs(logs []string) string {
	// TODO: Implement proper mint extraction from logs
	// For now, return empty string - will be improved with proper parsing
	return ""
}

func toJSON(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
