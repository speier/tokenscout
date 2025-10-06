package cli

import (
	"context"
	"fmt"

	"github.com/speier/tokenscout/internal/solana"
	"github.com/spf13/cobra"
)

var walletPath string

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet management commands",
	Long:  `Create, import, and manage Solana wallets.`,
}

var walletNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		wallet, err := solana.NewWallet()
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		if err := wallet.Save(walletPath); err != nil {
			return fmt.Errorf("failed to save wallet: %w", err)
		}

		fmt.Printf("✓ New wallet created and saved to: %s\n", walletPath)
		fmt.Printf("Public Address: %s\n", wallet.Address())
		fmt.Println("\n⚠️  IMPORTANT: Keep this file secure and back it up!")
		fmt.Println("⚠️  Anyone with access to this file can control your funds!")

		return nil
	},
}

var walletShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show wallet address and balance",
	RunE: func(cmd *cobra.Command, args []string) error {
		wallet, err := solana.LoadWallet(walletPath)
		if err != nil {
			return fmt.Errorf("failed to load wallet: %w", err)
		}

		rpcURL := cmd.Flag("rpc").Value.String()
		if rpcURL == "" {
			rpcURL = "https://api.mainnet-beta.solana.com"
		}

		client := solana.NewClient(rpcURL, wallet)
		ctx := context.Background()

		balance, err := client.GetBalance(ctx)
		if err != nil {
			return fmt.Errorf("failed to get balance: %w", err)
		}

		solBalance := float64(balance) / 1e9

		fmt.Printf("Wallet Address: %s\n", wallet.Address())
		fmt.Printf("Balance: %.4f SOL\n", solBalance)

		return nil
	},
}

var walletImportCmd = &cobra.Command{
	Use:   "import [source-file]",
	Short: "Import an existing wallet",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceFile := args[0]

		wallet, err := solana.LoadWallet(sourceFile)
		if err != nil {
			return fmt.Errorf("failed to load wallet from %s: %w", sourceFile, err)
		}

		if err := wallet.Save(walletPath); err != nil {
			return fmt.Errorf("failed to save wallet: %w", err)
		}

		fmt.Printf("✓ Wallet imported to: %s\n", walletPath)
		fmt.Printf("Public Address: %s\n", wallet.Address())

		return nil
	},
}

func init() {
	walletCmd.PersistentFlags().StringVar(&walletPath, "path", "wallet.json", "wallet file path")
	walletShowCmd.Flags().String("rpc", "https://api.mainnet-beta.solana.com", "Solana RPC URL")

	walletCmd.AddCommand(walletNewCmd)
	walletCmd.AddCommand(walletShowCmd)
	walletCmd.AddCommand(walletImportCmd)

	rootCmd.AddCommand(walletCmd)
}
