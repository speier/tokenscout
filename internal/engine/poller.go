package engine

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
)

// Poller polls RPC endpoints instead of using WebSocket
// Works with free public RPC that doesn't support WebSocket
type Poller struct {
	rpcClient *rpc.Client
	programs  []solana.PublicKey
	repo      repository.Repository
	eventCh   chan *models.Event
	interval  time.Duration
}

func NewPoller(rpcURL string, programIDs []string, repo repository.Repository, interval time.Duration) (*Poller, error) {
	programs := make([]solana.PublicKey, 0, len(programIDs))
	for _, id := range programIDs {
		pubkey, err := solana.PublicKeyFromBase58(id)
		if err != nil {
			logger.Error().Err(err).Str("program", id).Msg("Invalid program ID")
			continue
		}
		programs = append(programs, pubkey)
	}

	return &Poller{
		rpcClient: rpc.New(rpcURL),
		programs:  programs,
		repo:      repo,
		eventCh:   make(chan *models.Event, 100),
		interval:  interval,
	}, nil
}

func (p *Poller) Start(ctx context.Context) error {
	logger.Info().
		Int("programs", len(p.programs)).
		Dur("interval", p.interval).
		Msg("Starting RPC poller")

	// Track last processed signature for each program
	lastSigs := make(map[string]solana.Signature)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Poller shutting down")
			return nil
		case <-ticker.C:
			for _, program := range p.programs {
				if err := p.pollProgram(ctx, program, lastSigs); err != nil {
					logger.Error().
						Err(err).
						Str("program", program.String()).
						Msg("Failed to poll program")
				}
			}
		}
	}
}

func (p *Poller) pollProgram(ctx context.Context, program solana.PublicKey, lastSigs map[string]solana.Signature) error {
	// Get recent signatures for this program
	limit := 10
	opts := &rpc.GetSignaturesForAddressOpts{
		Limit:      &limit, // Check last 10 transactions
		Commitment: rpc.CommitmentFinalized,
	}

	// If we've seen signatures before, only get newer ones
	if lastSig, exists := lastSigs[program.String()]; exists {
		opts.Until = lastSig
	}

	sigs, err := p.rpcClient.GetSignaturesForAddressWithOpts(ctx, program, opts)
	if err != nil {
		return err
	}

	if len(sigs) == 0 {
		return nil
	}

	// Update last seen signature
	lastSigs[program.String()] = sigs[0].Signature

	// Process new signatures (in reverse order - oldest first)
	for i := len(sigs) - 1; i >= 0; i-- {
		sig := sigs[i]

		// Skip failed transactions
		if sig.Err != nil {
			continue
		}

		// Fetch transaction details
		tx, err := p.rpcClient.GetTransaction(ctx, sig.Signature, &rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentFinalized,
		})
		if err != nil {
			logger.Debug().
				Err(err).
				Str("signature", sig.Signature.String()).
				Msg("Failed to get transaction")
			continue
		}

		// Parse transaction for events
		event := p.parseTransaction(tx, program)
		if event != nil {
			// Store in database
			if err := p.repo.CreateEvent(ctx, event); err != nil {
				logger.Error().
					Err(err).
					Str("signature", sig.Signature.String()).
					Msg("Failed to store event")
				continue
			}

			// Send to event channel (removed duplicate log - already logged with 🔔)
			select {
			case p.eventCh <- event:
			default:
				logger.Warn().Msg("Event channel full, dropping event")
			}
		}
	}

	return nil
}

func (p *Poller) parseTransaction(tx *rpc.GetTransactionResult, program solana.PublicKey) *models.Event {
	if tx == nil || tx.Meta == nil {
		return nil
	}

	// Look for keywords in log messages
	logs := tx.Meta.LogMessages

	for _, log := range logs {
		// Detect new pool initialization
		if containsAny(log, []string{"InitializePool", "initialize", "CreatePool"}) {
			return &models.Event{
				Type:      models.EventTypeNewPool,
				Mint:      extractMintFromTransaction(tx),
				Timestamp: time.Now(),
				Raw:       toJSON(tx),
			}
		}

		// Detect new token mint
		if containsAny(log, []string{"InitializeMint", "CreateMint"}) {
			return &models.Event{
				Type:      models.EventTypeNewMint,
				Mint:      extractMintFromTransaction(tx),
				Timestamp: time.Now(),
				Raw:       toJSON(tx),
			}
		}
	}

	return nil
}

func extractMintFromTransaction(tx *rpc.GetTransactionResult) string {
	// TODO: Parse transaction accounts to find token mint
	// For now, return empty string
	return ""
}

func (p *Poller) EventChannel() <-chan *models.Event {
	return p.eventCh
}
