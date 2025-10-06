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

var positionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "Show open positions",
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

		positions, err := eng.GetPositions(ctx)
		if err != nil {
			return fmt.Errorf("failed to get positions: %w", err)
		}

		if len(positions) == 0 {
			fmt.Println("No open positions")
			return nil
		}

		output, err := json.MarshalIndent(positions, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(positionsCmd)
}
