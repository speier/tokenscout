package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
)

// WebhookListener receives Helius webhook events
type WebhookListener struct {
	port     int
	path     string
	secret   string
	repo     repository.Repository
	eventCh  chan *models.Event
	server   *http.Server
	parsers  *ParsersRegistry
}

func NewWebhookListener(port int, path string, secret string, repo repository.Repository) *WebhookListener {
	return &WebhookListener{
		port:    port,
		path:    path,
		secret:  secret,
		repo:    repo,
		eventCh: make(chan *models.Event, 100),
		parsers: NewParsersRegistry(),
	}
}

func (w *WebhookListener) Start(ctx context.Context) error {
	logger.Info().
		Int("port", w.port).
		Str("path", w.path).
		Msg("🌐 Starting webhook server...")

	mux := http.NewServeMux()
	mux.HandleFunc(w.path, w.handleWebhook)

	w.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", w.port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Webhook server error")
		}
	}()

	logger.Info().
		Int("port", w.port).
		Str("path", w.path).
		Msg("✅ Webhook server started")

	// Wait for context cancellation
	<-ctx.Done()
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return w.server.Shutdown(shutdownCtx)
}

func (w *WebhookListener) EventChannel() <-chan *models.Event {
	return w.eventCh
}

func (w *WebhookListener) handleWebhook(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse webhook payload
	var payload HeliusWebhookPayload
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		logger.Error().Err(err).Msg("Failed to parse webhook payload")
		http.Error(rw, "Bad request", http.StatusBadRequest)
		return
	}

	logger.Debug().
		Str("type", payload.Type).
		Int("transactions", len(payload.Transactions)).
		Msg("Received webhook")

	// Process each transaction in the payload
	for _, tx := range payload.Transactions {
		event := w.parseWebhookTransaction(tx)
		if event != nil {
			// Send event to channel
			select {
			case w.eventCh <- event:
				logger.Info().
					Str("mint", formatMint(event.Mint)).
					Str("signature", tx.Signature).
					Msg("🔔 New token from webhook")
			case <-time.After(time.Second):
				logger.Warn().Msg("Event channel full, dropping webhook event")
			}
		}
	}

	// Acknowledge receipt
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(map[string]string{"status": "ok"})
}

func (w *WebhookListener) parseWebhookTransaction(tx HeliusTransaction) *models.Event {
	// Helius already provides parsed data, check for new token mints
	// Look for token mints in account data or instructions
	
	// Check if this is a DEX program transaction
	for _, accountKey := range tx.AccountKeys {
		// Check if this is Raydium or Orca program
		if accountKey == RaydiumAMMV4.String() || accountKey == OrcaWhirlpool.String() {
			// Try to extract mint from token balances or transfers
			mint := w.extractMintFromHeliusTransaction(tx)
			if mint != "" {
				return &models.Event{
					Type:      models.EventTypeNewPool,
					Mint:      mint,
					Timestamp: time.Now(),
					Raw:       toJSON(tx),
				}
			}
		}
	}
	
	return nil
}

func (w *WebhookListener) extractMintFromHeliusTransaction(tx HeliusTransaction) string {
	// Helius provides parsed token transfers
	for _, transfer := range tx.TokenTransfers {
		// Look for non-SOL token mints
		if transfer.Mint != "" && transfer.Mint != WrappedSOL.String() {
			return transfer.Mint
		}
	}
	
	// Fallback: check native transfers for token accounts
	for _, native := range tx.NativeTransfers {
		// Skip if it's just SOL transfer
		if native.FromUserAccount != "" {
			continue
		}
	}
	
	return ""
}

// HeliusWebhookPayload represents the webhook payload from Helius
type HeliusWebhookPayload struct {
	Type         string               `json:"type"`
	Transactions []HeliusTransaction  `json:"transactions"`
}

type HeliusTransaction struct {
	Signature      string                  `json:"signature"`
	Type           string                  `json:"type"`
	Source         string                  `json:"source"`
	Timestamp      int64                   `json:"timestamp"`
	Slot           int64                   `json:"slot"`
	Fee            int64                   `json:"fee"`
	FeePayer       string                  `json:"feePayer"`
	AccountKeys    []string                `json:"accountKeys"`
	TokenTransfers []HeliusTokenTransfer   `json:"tokenTransfers"`
	NativeTransfers []HeliusNativeTransfer `json:"nativeTransfers"`
}

type HeliusTokenTransfer struct {
	FromUserAccount string `json:"fromUserAccount"`
	ToUserAccount   string `json:"toUserAccount"`
	FromTokenAccount string `json:"fromTokenAccount"`
	ToTokenAccount  string `json:"toTokenAccount"`
	Mint            string `json:"mint"`
	TokenAmount     string `json:"tokenAmount"`
}

type HeliusNativeTransfer struct {
	FromUserAccount string `json:"fromUserAccount"`
	ToUserAccount   string `json:"toUserAccount"`
	Amount          int64  `json:"amount"`
}
