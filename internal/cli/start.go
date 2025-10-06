package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/speier/tokenscout/internal/config"
	"github.com/speier/tokenscout/internal/engine"
	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/spf13/cobra"
)

var dryRun bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the trading bot",
	Long:  `Start the trading engine to monitor blockchain events and execute trades.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		logger.Init(logLevel, true)
		
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		logger.Debug().
			Str("rpc_url", cfg.Solana.RPCURL).
			Str("ws_url", cfg.Solana.WSURL).
			Msg("Loaded Solana config")

		if dryRun {
			cfg.Engine.Mode = "dry_run"
			logger.Info().Msg("Running in DRY RUN mode - no trades will be executed")
		}

		repo, err := repository.NewSQLite(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
		defer repo.Close()

		eng := engine.New(repo, cfg)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-sigChan
			fmt.Println("\nShutting down gracefully...")
			cancel()
		}()

		if err := eng.Start(ctx); err != nil {
			return fmt.Errorf("engine error: %w", err)
		}

		return nil
	},
}

func init() {
	startCmd.Flags().BoolVar(&dryRun, "dry-run", false, "run without executing trades")
	rootCmd.AddCommand(startCmd)
}
