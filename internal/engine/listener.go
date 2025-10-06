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
	wsURL     string
	programs  []solana.PublicKey
	repo      repository.Repository
	eventCh   chan *models.Event
	rpcClient *rpc.Client
	parsers   *ParsersRegistry
}

func NewListener(wsURL string, rpcURL string, programIDs []string, repo repository.Repository) (*Listener, error) {
	programs := make([]solana.PublicKey, 0, len(programIDs))
	for _, id := range programIDs {
		pubkey, err := solana.PublicKeyFromBase58(id)
		if err != nil {
			return nil, fmt.Errorf("invalid program ID %s: %w", id, err)
		}
		programs = append(programs, pubkey)
	}

	return &Listener{
		wsURL:     wsURL,
		programs:  programs,
		repo:      repo,
		eventCh:   make(chan *models.Event, 100),
		rpcClient: rpc.New(rpcURL),
		parsers:   NewParsersRegistry(),
	}, nil
}

func (l *Listener) Start(ctx context.Context) error {
	logger.Info().Msg("📡 Connecting to Solana blockchain...")
	logger.Debug().
		Str("ws_url", l.wsURL).
		Int("programs", len(l.programs)).
		Msg("WebSocket connection details")

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

	logger.Info().Msg("✅ Connected! Monitoring for new tokens...")
	logger.Debug().
		Int("programs", len(l.programs)).
		Msg("Subscribed to DEX programs")

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
	signature := logResult.Value.Signature
	
	// Fetch full transaction to parse instructions
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	maxVersion := uint64(0)
	tx, err := l.rpcClient.GetTransaction(
		ctx,
		signature,
		&rpc.GetTransactionOpts{
			Encoding:                       solana.EncodingBase64,
			MaxSupportedTransactionVersion: &maxVersion,
		},
	)
	
	if err != nil {
		// Only log non-rate-limit errors
		if !contains(err.Error(), "429") && !contains(err.Error(), "Too many") {
			logger.Debug().
				Err(err).
				Str("signature", signature.String()).
				Msg("Failed to fetch transaction")
		}
		return nil
	}
	
	if tx == nil || tx.Transaction == nil {
		return nil
	}
	
	// Parse transaction to extract mint addresses
	parsed, err := tx.Transaction.GetTransaction()
	if err != nil {
		logger.Debug().
			Err(err).
			Str("signature", signature.String()).
			Msg("Failed to parse transaction")
		return nil
	}
	
	// Use modular parsers to extract mint from transaction
	mint, dexName, found := l.extractMintFromTransaction(parsed, program)
	if !found {
		return nil
	}
	
	logger.Info().
		Str("mint", formatMint(mint)).
		Str("dex", dexName).
		Msg("🔔 New token detected")
	
	return &models.Event{
		Type:      models.EventTypeNewPool,
		Mint:      mint,
		Timestamp: time.Now(),
		Raw:       toJSON(logResult),
	}
}

func (l *Listener) extractMintFromTransaction(tx *solana.Transaction, program solana.PublicKey) (string, string, bool) {
	// Iterate through all instructions in the transaction
	for _, instruction := range tx.Message.Instructions {
		programID := tx.Message.AccountKeys[instruction.ProgramIDIndex]
		
		// Skip if not the program we're monitoring
		if !programID.Equals(program) {
			continue
		}
		
		// Get instruction accounts
		accounts := make([]solana.PublicKey, len(instruction.Accounts))
		for i, accountIndex := range instruction.Accounts {
			accounts[i] = tx.Message.AccountKeys[accountIndex]
		}
		
		// Use parsers registry to extract mint
		if mint, dexName, found := l.parsers.ParseInstruction(programID, accounts, instruction.Data); found {
			return mint, dexName, true
		}
	}
	
	return "", "", false
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
