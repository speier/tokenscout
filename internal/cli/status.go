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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show bot status and stats",
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

		stats, err := eng.GetStats(ctx)
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}

		output, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(output))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
