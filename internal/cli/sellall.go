package cli

import (
	"context"
	"fmt"

	"github.com/speier/tokenscout/internal/config"
	"github.com/speier/tokenscout/internal/engine"
	"github.com/speier/tokenscout/internal/logger"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/speier/tokenscout/internal/solana"
	"github.com/spf13/cobra"
)

var sellAllCmd = &cobra.Command{
	Use:   "sellall",
	Short: "Emergency: Close all open positions",
	Long:  `Immediately sell all open positions. Use in emergency situations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(logLevel, true)
		
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		repo, err := repository.NewSQLite(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
		defer repo.Close()

		// Load wallet
		wallet, err := solana.LoadWallet(cfg.Solana.WalletPath)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		// Create clients
		solanaClient := solana.NewClient(cfg.Solana.RPCURL, wallet)
		jupiterClient := solana.NewJupiterClient(cfg.Solana.JupiterAPIURL)

		// Create executor
		executor := engine.NewExecutor(cfg, repo, solanaClient, jupiterClient)

		// Get positions
		ctx := context.Background()
		positions, err := repo.GetAllPositions(ctx)
		if err != nil {
			return fmt.Errorf("failed to get positions: %w", err)
		}

		if len(positions) == 0 {
			fmt.Println("No open positions to sell")
			return nil
		}

		fmt.Printf("Found %d open position(s)\n", len(positions))
		for _, pos := range positions {
			fmt.Printf("  - %s (qty: %s)\n", pos.Mint, pos.Quantity)
		}

		// Confirm
		fmt.Print("\nAre you sure you want to sell all positions? (yes/no): ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" {
			fmt.Println("Cancelled")
			return nil
		}

		// Execute sell all
		logger.Info().Int("count", len(positions)).Msg("Selling all positions")
		if err := executor.SellAll(ctx, "manual_sellall"); err != nil {
			return fmt.Errorf("failed to sell all: %w", err)
		}

		fmt.Println("All positions closed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sellAllCmd)
}
