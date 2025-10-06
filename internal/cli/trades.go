package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/speier/tokenscout/internal/config"
	"github.com/speier/tokenscout/internal/engine"
	"github.com/speier/tokenscout/internal/repository"
	"github.com/spf13/cobra"
)

var tradesLimit int

var tradesCmd = &cobra.Command{
	Use:   "trades",
	Short: "Show recent trades",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		repo, err := repository.NewSQLite(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
		defer repo.Close()

		eng := engine.New(repo, cfg)
		ctx := context.Background()

		trades, err := eng.GetRecentTrades(ctx, tradesLimit)
		if err != nil {
			return fmt.Errorf("failed to get trades: %w", err)
		}

		output, err := json.MarshalIndent(trades, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	tradesCmd.Flags().IntVarP(&tradesLimit, "limit", "l", 20, "number of trades to show")
	rootCmd.AddCommand(tradesCmd)
}
