package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
)

// WatchedToken represents a token that was rejected but might become valid
type WatchedToken struct {
	Mint          string
	Event         *models.Event
	FirstSeenAt   time.Time
	LastCheckedAt time.Time
	RejectReason  string
	CheckCount    int
}

type Processor struct {
	eventCh   <-chan *models.Event
	engine    *engine
	executor  *Executor
	watchList map[string]*WatchedToken
	watchMux  sync.RWMutex

	// Stats tracking for periodic summaries
	stats    *processorStats
	statsMux sync.Mutex

	// Rolling log for recent events
	recentEvents []eventLog
	eventMux     sync.Mutex
}

type eventLog struct {
	timestamp time.Time
	eventType string // "reject", "watch", "watch_success", "watch_expired", "buy", "buy_fail"
	mint      string
	message   string
}

type processorStats struct {
	tokensDetected   int
	tokensRejected   int
	tokensBought     int
	rejectionReasons map[string]int
	lastReset        time.Time
}

func NewProcessor(eventCh <-chan *models.Event, eng *engine, executor *Executor) *Processor {
	return &Processor{
		eventCh:   eventCh,
		engine:    eng,
		executor:  executor,
		watchList: make(map[string]*WatchedToken),
		stats: &processorStats{
			rejectionReasons: make(map[string]int),
			lastReset:        time.Now(),
		},
	}
}

func (p *Processor) Start(ctx context.Context) error {
	logger.Info().Msg("⚙️ Starting event processor")

	// Track recently seen mints to prevent duplicate processing
	// Use longer window to prevent re-processing same pools
	seenMints := make(map[string]time.Time)
	dedupeWindow := 5 * time.Minute // Don't reprocess same mint for 5 minutes

	cleanupTicker := time.NewTicker(1 * time.Minute) // Cleanup every minute
	defer cleanupTicker.Stop()

	recheckTicker := time.NewTicker(15 * time.Second) // Re-check watched tokens every 15s
	defer recheckTicker.Stop()

	summaryTicker := time.NewTicker(10 * time.Second) // Print summary every 10s
	defer summaryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Event processor shutting down")
			return nil

		case event := <-p.eventCh:
			// Deduplicate based on mint address
			if lastSeen, exists := seenMints[event.Mint]; exists {
				if time.Since(lastSeen) < dedupeWindow {
					// Skip silently - don't log duplicate spam
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

		case <-recheckTicker.C:
			// Re-evaluate watched tokens
			p.recheckWatchedTokens(ctx)

		case <-summaryTicker.C:
			// Print periodic summary
			p.printStatusLine()

		case <-cleanupTicker.C:
			// Clean up old entries from seenMints (older than 10 minutes)
			cutoff := time.Now().Add(-10 * time.Minute)
			for mint, t := range seenMints {
				if t.Before(cutoff) {
					delete(seenMints, mint)
					// Don't log cleanup - too verbose
				}
			}

			// Clean up expired watch list entries (older than 2 minutes)
			p.cleanupWatchList()
		}
	}
}

func (p *Processor) processEvent(ctx context.Context, event *models.Event) error {
	// Track stats
	p.statsMux.Lock()
	p.stats.tokensDetected++
	p.statsMux.Unlock()

	// Evaluate rules
	ruleEngine := NewRuleEngine(p.engine.config, p.engine.repo, p.engine.config.Solana.RPCURL)
	decision, err := ruleEngine.Evaluate(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to evaluate rules: %w", err)
	}

	if !decision.Allow {
		reason := decision.Reasons[0]

		// Add to watch list if rejection is temporary (might change)
		if p.isWatchableRejection(reason) {
			p.addToWatchList(event, reason)
			// Don't count as rejected yet - we're giving it a chance
			// Watch list addition is logged in addToWatchList
		} else {
			// Only log and count permanent rejections
			p.addEventToLog("reject", event.Mint, reason)
			p.statsMux.Lock()
			p.stats.tokensRejected++
			p.stats.rejectionReasons[reason]++
			p.statsMux.Unlock()
		}

		return nil
	}

	// Remove from watch list if it was being watched
	p.removeFromWatchList(event.Mint)

	// Token passes rules - this is important, log it!
	p.clearStatusDisplay() // Clear the rolling display
	logger.Info().
		Str("mint", formatMint(event.Mint)).
		Msg("✅ Token passes rules, executing buy")

	p.statsMux.Lock()
	p.stats.tokensBought++
	p.statsMux.Unlock()

	if err := p.executor.ExecuteBuy(ctx, event.Mint, "rules_passed"); err != nil {
		p.clearStatusDisplay() // Clear the rolling display
		logger.Error().
			Err(err).
			Str("mint", event.Mint).
			Msg("Failed to execute buy")
		p.addEventToLog("buy_fail", event.Mint, fmt.Sprintf("error: %v", err))
		return fmt.Errorf("failed to execute buy: %w", err)
	}

	// Log successful buy
	p.addEventToLog("buy", event.Mint, "rules_passed")

	return nil
}

// isWatchableRejection determines if a rejection reason is temporary and worth re-checking
func (p *Processor) isWatchableRejection(reason string) bool {
	// Watch these - they can change over time
	watchableReasons := []string{
		"holders:",   // Holder count can increase
		"liquidity:", // Liquidity can increase
		"mint age:",  // Token gets older (might become valid)
	}

	// Don't watch these - they're permanent rejections
	// - freeze authority
	// - mint authority
	// - blacklisted
	// - failed to fetch token info (likely invalid/malformed)

	for _, watchable := range watchableReasons {
		if containsIgnoreCase(reason, watchable) {
			return true
		}
	}
	return false
}

// addToWatchList adds a token to the watch list for re-evaluation
func (p *Processor) addToWatchList(event *models.Event, reason string) {
	p.watchMux.Lock()
	defer p.watchMux.Unlock()

	// Don't re-add if already watching
	if _, exists := p.watchList[event.Mint]; exists {
		return
	}

	p.watchList[event.Mint] = &WatchedToken{
		Mint:          event.Mint,
		Event:         event,
		FirstSeenAt:   time.Now(),
		LastCheckedAt: time.Now(),
		RejectReason:  reason,
		CheckCount:    1,
	}

	// Log to rolling log
	p.addEventToLog("watch", event.Mint, reason)
}

// removeFromWatchList removes a token from the watch list
func (p *Processor) removeFromWatchList(mint string) {
	p.watchMux.Lock()
	defer p.watchMux.Unlock()
	delete(p.watchList, mint)
}

// recheckWatchedTokens re-evaluates all tokens in the watch list
func (p *Processor) recheckWatchedTokens(ctx context.Context) {
	p.watchMux.RLock()
	tokens := make([]*WatchedToken, 0, len(p.watchList))
	for _, token := range p.watchList {
		tokens = append(tokens, token)
	}
	p.watchMux.RUnlock()

	if len(tokens) == 0 {
		return
	}

	// Don't log every re-check, only interesting outcomes

	for _, token := range tokens {
		// Re-evaluate the token
		ruleEngine := NewRuleEngine(p.engine.config, p.engine.repo, p.engine.config.Solana.RPCURL)
		decision, err := ruleEngine.Evaluate(ctx, token.Event)

		p.watchMux.Lock()
		token.LastCheckedAt = time.Now()
		token.CheckCount++
		p.watchMux.Unlock()

		if err != nil {
			// Only log errors, not routine failures
			continue
		}

		if decision.Allow {
			watchTime := time.Since(token.FirstSeenAt)

			// Token now passes rules! This is important - log it
			p.clearStatusDisplay() // Clear the rolling display
			logger.Info().
				Str("mint", formatMint(token.Mint)).
				Int("checks", token.CheckCount).
				Dur("watch_time", watchTime).
				Msg("✅ Watch list success! Token now passes rules")

			p.removeFromWatchList(token.Mint)

			// Log watch success
			p.addEventToLog("watch_success", token.Mint,
				fmt.Sprintf("passed after %s", formatDuration(watchTime)))

			p.statsMux.Lock()
			p.stats.tokensBought++
			p.statsMux.Unlock()

			if err := p.executor.ExecuteBuy(ctx, token.Mint, "rules_passed_after_watch"); err != nil {
				p.clearStatusDisplay() // Clear the rolling display
				logger.Error().
					Err(err).
					Str("mint", token.Mint).
					Msg("Failed to execute buy for watched token")
				p.addEventToLog("buy_fail", token.Mint, fmt.Sprintf("error: %v", err))
			} else {
				p.addEventToLog("buy", token.Mint, "rules_passed_after_watch")
			}
		}
		// Don't log "still rejected" - it's noise
	}
}

// cleanupWatchList removes tokens that have been watched too long (2 min max)
func (p *Processor) cleanupWatchList() {
	p.watchMux.Lock()
	defer p.watchMux.Unlock()

	maxWatchDuration := 2 * time.Minute
	now := time.Now()
	expired := []string{}
	expiredTokens := []*WatchedToken{}

	for mint, token := range p.watchList {
		if now.Sub(token.FirstSeenAt) > maxWatchDuration {
			expired = append(expired, mint)
			expiredTokens = append(expiredTokens, token)
		}
	}

	// Only log if we actually removed something
	if len(expired) > 0 {
		logger.Debug().
			Int("count", len(expired)).
			Msg("Cleaned up expired tokens from watch list")

		// Count expired tokens as rejected now (they didn't pass rules)
		p.statsMux.Lock()
		p.stats.tokensRejected += len(expired)
		p.statsMux.Unlock()

		// Log expired tokens
		for _, token := range expiredTokens {
			p.addEventToLog("watch_expired", token.Mint, token.RejectReason)
		}
	}

	for _, mint := range expired {
		delete(p.watchList, mint)
	}
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && indexIgnoreCase(s, substr) >= 0
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func indexIgnoreCase(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0
	}
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}

// addEventToLog adds an event to the rolling log (keeps last 5)
func (p *Processor) addEventToLog(eventType, mint, message string) {
	p.eventMux.Lock()
	defer p.eventMux.Unlock()

	// Add new event
	p.recentEvents = append(p.recentEvents, eventLog{
		timestamp: time.Now(),
		eventType: eventType,
		mint:      mint,
		message:   message,
	})

	// Keep only last 5 events
	if len(p.recentEvents) > 5 {
		p.recentEvents = p.recentEvents[len(p.recentEvents)-5:]
	}
	// Don't print here - let the status ticker handle it
}

// clearStatusDisplay clears the display before important messages
func (p *Processor) clearStatusDisplay() {
	// Simple newline - let logs scroll naturally
	fmt.Println()
}

// printStatusLine prints a periodic summary block
func (p *Processor) printStatusLine() {
	p.statsMux.Lock()
	detected := p.stats.tokensDetected
	rejected := p.stats.tokensRejected
	bought := p.stats.tokensBought
	p.statsMux.Unlock()

	p.watchMux.RLock()
	watching := len(p.watchList)
	p.watchMux.RUnlock()

	// Get recent events
	p.eventMux.Lock()
	recentEvents := make([]eventLog, len(p.recentEvents))
	copy(recentEvents, p.recentEvents)
	p.eventMux.Unlock()

	// Print timestamp
	now := time.Now().Format("15:04:05")
	fmt.Printf("\n[%s] Recent Activity:\n", now)

	// Print up to 5 recent events
	if len(recentEvents) == 0 {
		fmt.Println("  (No events yet)")
	} else {
		for _, event := range recentEvents {
			elapsed := time.Since(event.timestamp)
			icon := getEventIcon(event.eventType)
			fmt.Printf("  %s %s | %s | %s ago\n",
				icon,
				formatMint(event.mint),
				event.message,
				formatDuration(elapsed))
		}
	}

	// Print status line
	fmt.Printf("\n📊 %d detected | %d rejected | %d watching | %d bought\n",
		detected, rejected, watching, bought)
	fmt.Println("─────────────────────────────────────────────────────────")
} // getEventIcon returns the appropriate icon for an event type
func getEventIcon(eventType string) string {
	switch eventType {
	case "reject":
		return "❌"
	case "watch":
		return "🔍"
	case "watch_success":
		return "🎯"
	case "watch_expired":
		return "⏱️"
	case "buy":
		return "✅"
	case "buy_fail":
		return "⚠️"
	default:
		return "•"
	}
}

// formatDuration formats a duration in a human-readable short form
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "just now"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
