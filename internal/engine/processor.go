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

	// Coalescing window to deduplicate events
	coalesceWindow := time.Duration(p.engine.config.Listener.CoalesceWindowMs) * time.Millisecond
	seenMints := make(map[string]time.Time)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Event processor shutting down")
			return nil

		case event := <-p.eventCh:
			// Deduplicate events within coalesce window
			if lastSeen, exists := seenMints[event.Mint]; exists {
				if time.Since(lastSeen) < coalesceWindow {
					logger.Debug().
						Str("mint", event.Mint).
						Msg("Ignoring duplicate event within coalesce window")
					continue
				}
			}

			seenMints[event.Mint] = time.Now()

			// Process the event
			if err := p.processEvent(ctx, event); err != nil {
				logger.Error().
					Err(err).
					Str("mint", event.Mint).
					Msg("Failed to process event")
			}

		case <-ticker.C:
			// Clean up old entries from seenMints
			cutoff := time.Now().Add(-coalesceWindow * 10)
			for mint, t := range seenMints {
				if t.Before(cutoff) {
					delete(seenMints, mint)
				}
			}
		}
	}
}

func (p *Processor) processEvent(ctx context.Context, event *models.Event) error {
	logger.Info().
		Str("type", string(event.Type)).
		Str("mint", event.Mint).
		Msg("Processing event")

	// Evaluate rules
	ruleEngine := NewRuleEngine(p.engine.config, p.engine.repo, p.engine.config.Solana.RPCURL)
	decision, err := ruleEngine.Evaluate(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to evaluate rules: %w", err)
	}

	if !decision.Allow {
		logger.Info().
			Str("mint", event.Mint).
			Strs("reasons", decision.Reasons).
			Msg("Token rejected by rules")
		return nil
	}

	// Token passed rules - execute buy
	logger.Info().
		Str("mint", event.Mint).
		Msg("Token passed rule evaluation, executing buy")

	if err := p.executor.ExecuteBuy(ctx, event.Mint, "rules_passed"); err != nil {
		logger.Error().
			Err(err).
			Str("mint", event.Mint).
			Msg("Failed to execute buy")
		return fmt.Errorf("failed to execute buy: %w", err)
	}

	return nil
}
