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
	"github.com/speier/tokenscout/internal/strategies"
	"github.com/spf13/cobra"
)

var (
	dryRun         bool
	strategyName   string
	listStrategies bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the trading bot",
	Long:  `Start the trading engine to monitor blockchain events and execute trades.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If --list-strategies, show available strategies and exit
		if listStrategies {
			fmt.Println("\n═══════════════════════════════════════════════════════════════════════")
			fmt.Println("📋 Available Strategy Presets")
			fmt.Println("═══════════════════════════════════════════════════════════════════════")
			fmt.Println()
			fmt.Printf("%-18s %-13s %-14s %-13s %-8s\n", "Strategy", "Hold Time", "Entry", "Exit", "Risk")
			fmt.Println("-----------------------------------------------------------------------")

			for _, info := range strategies.GetStrategyInfo() {
				fmt.Printf("%-18s %-13s %-14s %-13s %-8s\n",
					info.Name,
					info.HoldTime,
					info.Entry,
					info.Exit,
					info.Risk)
			}

			fmt.Println()
			fmt.Println("Detailed descriptions:")
			for _, desc := range strategies.ListStrategies() {
				fmt.Println(desc)
			}
			fmt.Println("\n═══════════════════════════════════════════════════════════════════════")
			fmt.Println("Usage:")
			fmt.Println("  tokenscout start --strategy <name>")
			fmt.Println("  tokenscout start --strategy snipe_flip --dry-run")
			fmt.Println("═══════════════════════════════════════════════════════════════════════\n")
			return nil
		}

		// Initialize logger
		logger.Init(logLevel, true)

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Apply strategy preset if specified
		if strategyName != "" {
			logger.Get().Info().
				Str("strategy", strategyName).
				Msg("📋 Applying strategy preset")

			cfg, err = strategies.ApplyStrategy(cfg, strategyName)
			if err != nil {
				return fmt.Errorf("failed to apply strategy: %w", err)
			}

			// Set the strategy name in the config
			cfg.Strategy = strategyName

			// Log strategy details
			strategy, _ := strategies.GetStrategy(strategyName)
			logger.Get().Info().
				Str("description", strategy.Description).
				Msg("Strategy configuration loaded")
		} else {
			// No strategy specified, use "custom" as default
			cfg.Strategy = "custom"
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
	startCmd.Flags().StringVar(&strategyName, "strategy", "", "strategy preset (snipe_flip, conservative, scalping, data_collection, momentum_rider)")
	startCmd.Flags().BoolVar(&listStrategies, "list-strategies", false, "list all available strategy presets")
	rootCmd.AddCommand(startCmd)
}
