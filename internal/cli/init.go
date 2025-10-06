package cli

import (
	"fmt"

	"github.com/speier/tokenscout/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  `Create a default config.yaml file with sensible defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.CreateDefault(cfgFile); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}

		fmt.Printf("Configuration file created: %s\n", cfgFile)
		fmt.Println("Edit this file to customize your trading rules and settings.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
