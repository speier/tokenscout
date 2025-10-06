package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
)

type Processor struct {
	eventCh  <-chan *models.Event
	engine   *engine
	executor *Executor
}

func NewProcessor(eventCh <-chan *models.Event, eng *engine, executor *Executor) *Processor {
	return &Processor{
		eventCh:  eventCh,
		engine:   eng,
		executor: executor,
	}
}

func (p *Processor) Start(ctx context.Context) error {
	logger.Info().Msg("Starting event processor")

	// Track recently seen mints to prevent duplicate processing
	// Use longer window to prevent re-processing same pools
	seenMints := make(map[string]time.Time)
	dedupeWindow := 5 * time.Minute // Don't reprocess same mint for 5 minutes

	ticker := time.NewTicker(1 * time.Minute) // Cleanup every minute
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Event processor shutting down")
			return nil

		case event := <-p.eventCh:
			// Deduplicate based on mint address
			if lastSeen, exists := seenMints[event.Mint]; exists {
				if time.Since(lastSeen) < dedupeWindow {
					logger.Debug().
						Str("mint", formatMint(event.Mint)).
						Dur("since_last", time.Since(lastSeen)).
						Msg("Skipping duplicate event (already processed)")
					continue
				}
			}

			// Mark as seen
			seenMints[event.Mint] = time.Now()

			// Process the event
			if err := p.processEvent(ctx, event); err != nil {
				logger.Error().
					Err(err).
					Str("mint", event.Mint).
					Msg("Failed to process event")
			}

		case <-ticker.C:
			// Clean up old entries from seenMints (older than 10 minutes)
			cutoff := time.Now().Add(-10 * time.Minute)
			for mint, t := range seenMints {
				if t.Before(cutoff) {
					delete(seenMints, mint)
					logger.Debug().
						Str("mint", formatMint(mint)).
						Msg("Removed from deduplication cache")
				}
			}
		}
	}
}

func (p *Processor) processEvent(ctx context.Context, event *models.Event) error {
	shortMint := event.Mint
	if len(event.Mint) > 8 {
		shortMint = event.Mint[:4] + ".." + event.Mint[len(event.Mint)-4:]
	}

	logger.Info().
		Str("mint", shortMint).
		Msg("→ Evaluating token")

	// Evaluate rules
	ruleEngine := NewRuleEngine(p.engine.config, p.engine.repo, p.engine.config.Solana.RPCURL)
	decision, err := ruleEngine.Evaluate(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to evaluate rules: %w", err)
	}

	if !decision.Allow {
		logger.Info().
			Str("mint", shortMint).
			Str("reason", decision.Reasons[0]).
			Msg("✗ Rejected")
		return nil
	}

	if err := p.executor.ExecuteBuy(ctx, event.Mint, "rules_passed"); err != nil {
		logger.Error().
			Err(err).
			Str("mint", event.Mint).
			Msg("Failed to execute buy")
		return fmt.Errorf("failed to execute buy: %w", err)
	}

	return nil
}
