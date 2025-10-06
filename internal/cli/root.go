package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	dbPath   string
	logLevel string
)

var rootCmd = &cobra.Command{
	Use:   "tokenscout",
	Short: "TokenScout - Solana new-token trading bot",
	Long:  `A CLI-based trading bot that monitors newly created tokens on Solana and trades automatically based on rules.`,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "tokenscout.db", "database file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level (debug, info, warn, error)")
}
