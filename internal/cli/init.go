package cli

import (
	"fmt"
	"os"

	"github.com/speier/tokenscout/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration files",
	Long:  `Create default config.yaml and .env files with sensible defaults.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create config.yaml
		if err := config.CreateDefault(cfgFile); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		fmt.Printf("✓ Configuration file created: %s\n", cfgFile)

		// Create .env
		envFile := ".env"
		if err := config.CreateEnvTemplate(envFile); err != nil {
			if os.IsExist(err) || err.Error() == fmt.Sprintf("env file already exists: %s", envFile) {
				fmt.Printf("⚠ Environment file already exists: %s (skipping)\n", envFile)
			} else {
				return fmt.Errorf("failed to create env: %w", err)
			}
		} else {
			fmt.Printf("✓ Environment file created: %s\n", envFile)
		}

		fmt.Println("\nNext steps:")
		fmt.Println("1. Edit .env and add your Solana RPC URLs (required)")
		fmt.Println("2. Edit config.yaml to customize trading rules (optional)")
		fmt.Println("3. Run 'tokenscout wallet' to create or import a wallet")
		fmt.Println("4. Run 'tokenscout start' to begin trading")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
