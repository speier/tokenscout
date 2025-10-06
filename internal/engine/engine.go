package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/models"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/speier/tokenscout/internal/solana"
)

type Status struct {
	Running       bool   `json:"running"`
	Mode          string `json:"mode"`
	OpenPositions int    `json:"open_positions"`
	TotalTrades   int    `json:"total_trades"`
}

type Stats struct {
	Status        Status  `json:"status"`
	WalletBalance float64 `json:"wallet_balance"`
	TodayPnL      float64 `json:"today_pnl"`
}

type Engine interface {
	Start(ctx context.Context) error
	Stop() error
	Status() Status
	ExecuteTrade(ctx context.Context, trade *models.Trade) error
	GetPositions(ctx context.Context) ([]models.Position, error)
	ClosePosition(ctx context.Context, mint string) error
	CloseAllPositions(ctx context.Context) error
	GetConfig() *models.Config
	UpdateConfig(cfg *models.Config) error
	ReloadRules() error
	GetStats(ctx context.Context) (Stats, error)
	GetRecentTrades(ctx context.Context, limit int) ([]models.Trade, error)
}

type engine struct {
	repo      repository.Repository
	config    *models.Config
	status    Status
	mu        sync.RWMutex
	cancel    context.CancelFunc
	listener  *Listener
	processor *Processor
	executor  *Executor
	monitor   *Monitor
}

func New(repo repository.Repository, config *models.Config) Engine {
	return &engine{
		repo:   repo,
		config: config,
		status: Status{
			Running: false,
			Mode:    string(config.Engine.Mode),
		},
	}
}

func (e *engine) Start(ctx context.Context) error {
	e.mu.Lock()
	if e.status.Running {
		e.mu.Unlock()
		return fmt.Errorf("engine already running")
	}

	ctx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.status.Running = true
	e.mu.Unlock()

	if e.config.Engine.Mode == models.ModeDryRun {
		logger.Info().Msg("🤖 Bot started in simulation mode (no real money)")
	} else {
		logger.Info().Msg("🤖 Bot started in LIVE mode")
	}
	
	logger.Debug().
		Str("mode", string(e.config.Engine.Mode)).
		Int("max_positions", e.config.Trading.MaxOpenPositions).
		Msg("Engine configuration")

	// Initialize executor and monitor
	wallet, err := e.loadWallet()
	if err != nil {
		logger.Info().Msg("💰 No wallet loaded - will simulate trades only")
		logger.Debug().Err(err).Msg("Wallet load details")
	} else {
		solanaClient := solana.NewClient(e.config.Solana.RPCURL, wallet)
		jupiterClient := solana.NewJupiterClient(e.config.Solana.JupiterAPIURL)
		
		e.executor = NewExecutor(e.config, e.repo, solanaClient, jupiterClient)
		e.monitor = NewMonitor(e.config, e.repo, e.executor, jupiterClient)
		
		// Start position monitor
		go func() {
			if err := e.monitor.Start(ctx); err != nil {
				logger.Error().Err(err).Msg("Monitor error")
			}
		}()
	}

	// Start blockchain listener if enabled
	if e.config.Listener.Enabled {
		var eventCh <-chan *models.Event
		
		if e.config.Listener.Mode == "websocket" {
			// WebSocket mode (free with rate limits)
			logger.Debug().
				Str("ws_url", e.config.Solana.WSURL).
				Msg("Creating WebSocket listener")
			
			listener, err := NewListener(
				e.config.Solana.WSURL,
				e.config.Solana.RPCURL,
				e.config.Listener.Programs,
				e.repo,
			)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create WebSocket listener")
			} else {
				e.listener = listener
				eventCh = listener.EventChannel()
				
				go func() {
					if err := listener.Start(ctx); err != nil {
						logger.Error().Err(err).Msg("Listener error")
					}
				}()
			}
		} else {
			// Polling mode (works with free RPC)
			pollingInterval := e.config.Listener.PollingInterval
			if pollingInterval <= 0 {
				pollingInterval = 10 // Default to 10 seconds
			}
			interval := time.Duration(pollingInterval) * time.Second
			poller, err := NewPoller(
				e.config.Solana.RPCURL,
				e.config.Listener.Programs,
				e.repo,
				interval,
			)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to create poller")
			} else {
				eventCh = poller.EventChannel()
				
				go func() {
					if err := poller.Start(ctx); err != nil {
						logger.Error().Err(err).Msg("Poller error")
					}
				}()
			}
		}
		
		// Start event processor if we have an event channel
		if eventCh != nil && e.executor != nil {
			e.processor = NewProcessor(eventCh, e, e.executor)
			go func() {
				if err := e.processor.Start(ctx); err != nil {
					logger.Error().Err(err).Msg("Processor error")
				}
			}()
		}
	} else {
		logger.Info().Msg("Listener disabled in config")
	}
	
	<-ctx.Done()
	logger.Info().Msg("Trading engine shutting down")
	return nil
}

func (e *engine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.status.Running {
		return fmt.Errorf("engine not running")
	}

	if e.cancel != nil {
		e.cancel()
	}

	e.status.Running = false
	logger.Info().Msg("Trading engine stopped")
	return nil
}

func (e *engine) Status() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

func (e *engine) ExecuteTrade(ctx context.Context, trade *models.Trade) error {
	// TODO: Implement trade execution logic
	return e.repo.CreateTrade(ctx, trade)
}

func (e *engine) GetPositions(ctx context.Context) ([]models.Position, error) {
	return e.repo.GetAllPositions(ctx)
}

func (e *engine) ClosePosition(ctx context.Context, mint string) error {
	// TODO: Execute sell order
	return e.repo.DeletePosition(ctx, mint)
}

func (e *engine) CloseAllPositions(ctx context.Context) error {
	positions, err := e.repo.GetAllPositions(ctx)
	if err != nil {
		return err
	}

	for _, pos := range positions {
		if err := e.ClosePosition(ctx, pos.Mint); err != nil {
			return err
		}
	}
	return nil
}

func (e *engine) GetConfig() *models.Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

func (e *engine) UpdateConfig(cfg *models.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.config = cfg
	e.status.Mode = string(cfg.Engine.Mode)
	return nil
}

func (e *engine) ReloadRules() error {
	// TODO: Reload rules from config or database
	fmt.Println("Rules reloaded")
	return nil
}

func (e *engine) GetStats(ctx context.Context) (Stats, error) {
	positions, err := e.repo.GetAllPositions(ctx)
	if err != nil {
		return Stats{}, err
	}

	trades, err := e.repo.GetTrades(ctx, 100)
	if err != nil {
		return Stats{}, err
	}

	e.mu.RLock()
	status := e.status
	e.mu.RUnlock()

	status.OpenPositions = len(positions)
	status.TotalTrades = len(trades)

	return Stats{
		Status: status,
		// TODO: Calculate actual wallet balance and PnL
		WalletBalance: 0,
		TodayPnL:      0,
	}, nil
}

func (e *engine) GetRecentTrades(ctx context.Context, limit int) ([]models.Trade, error) {
	return e.repo.GetTrades(ctx, limit)
}

func (e *engine) loadWallet() (*solana.Wallet, error) {
	walletPath := e.config.Solana.WalletPath
	if walletPath == "" {
		walletPath = "wallet.json"
	}
	return solana.LoadWallet(walletPath)
}
