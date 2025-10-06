package cli

import (
	"context"
	"fmt"

	"github.com/speier/tokenscout/internal/repository"
	"github.com/spf13/cobra"
)

var strategiesCmd = &cobra.Command{
	Use:   "strategies",
	Short: "Strategy management and analytics",
	Long:  `Commands for managing and comparing trading strategies.`,
}

var strategiesCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare performance across strategies",
	Long:  `Display performance metrics for all strategies that have been used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := repository.NewSQLite(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize repository: %w", err)
		}
		defer repo.Close()

		ctx := context.Background()
		stats, err := repo.GetStrategyStats(ctx)
		if err != nil {
			return fmt.Errorf("failed to get strategy stats: %w", err)
		}

		if len(stats) == 0 {
			fmt.Println("No trading activity yet. Start trading to see strategy performance!")
			return nil
		}

		// Print header
		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")
		fmt.Println("📊 Strategy Performance Comparison")
		fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")
		fmt.Println()

		// Print table header
		fmt.Printf("%-18s %8s %8s %8s %8s %12s %12s %10s\n",
			"Strategy", "Trades", "Buy", "Sell", "Open", "Avg Entry", "Volume USD", "Success %")
		fmt.Println("-------------------------------------------------------------------------------------------")

		// Print each strategy
		for _, s := range stats {
			fmt.Printf("%-18s %8d %8d %8d %8d $%11.6f $%11.2f %9.1f%%\n",
				s.Strategy,
				s.TotalTrades,
				s.BuyTrades,
				s.SellTrades,
				s.OpenPositions,
				s.AvgEntryPrice,
				s.TotalVolume,
				s.SuccessRate,
			)
		}

		fmt.Println()
		fmt.Println("Notes:")
		fmt.Println("  • Success % = (Executed Trades / Total Trades) × 100")
		fmt.Println("  • Volume USD = Total value of buy trades executed")
		fmt.Println("  • Avg Entry = Average price paid when buying (USD per token)")
		fmt.Println("═══════════════════════════════════════════════════════════════════════════════════════")
		fmt.Println()

		return nil
	},
}

func init() {
	strategiesCmd.AddCommand(strategiesCompareCmd)
	rootCmd.AddCommand(strategiesCmd)
}
