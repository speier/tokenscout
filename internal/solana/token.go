package solana

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// TokenInfo represents information about a SPL token
type TokenInfo struct {
	Mint              string
	Name              string
	Symbol            string
	Decimals          uint8
	Supply            uint64
	HasFreezeAuthority bool
	HasMintAuthority   bool
	FreezeAuthority    *solana.PublicKey
	MintAuthority      *solana.PublicKey
}

// TokenAccountInfo represents holder information
type TokenAccountInfo struct {
	Owner   solana.PublicKey
	Amount  uint64
	Address solana.PublicKey
}

// GetTokenInfo fetches token metadata and mint information
func GetTokenInfo(ctx context.Context, client *rpc.Client, mintAddress string) (*TokenInfo, error) {
	mint, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid mint address: %w", err)
	}

	// Get mint account info
	accountInfo, err := client.GetAccountInfo(ctx, mint)
	if err != nil {
		return nil, fmt.Errorf("failed to get mint account: %w", err)
	}

	if accountInfo == nil || accountInfo.Value == nil {
		return nil, fmt.Errorf("mint account not found")
	}

	data := accountInfo.Value.Data.GetBinary()
	if len(data) < 82 {
		return nil, fmt.Errorf("invalid mint account data")
	}

	info := &TokenInfo{
		Mint: mintAddress,
	}

	// Parse mint account data (SPL Token format)
	// Offset 0-32: mint authority (optional)
	// Offset 36: decimals
	// Offset 40-44: is_initialized
	// Offset 44-76: freeze authority (optional)

	info.Decimals = data[44]

	// Check if mint authority exists
	mintAuthFlag := data[0]
	if mintAuthFlag == 1 {
		info.HasMintAuthority = true
		var mintAuth solana.PublicKey
		copy(mintAuth[:], data[4:36])
		info.MintAuthority = &mintAuth
	}

	// Check if freeze authority exists
	freezeAuthFlag := data[46]
	if freezeAuthFlag == 1 {
		info.HasFreezeAuthority = true
		var freezeAuth solana.PublicKey
		copy(freezeAuth[:], data[46:78])
		info.FreezeAuthority = &freezeAuth
	}

	// TODO: Fetch name and symbol from Metaplex metadata
	// This requires querying the metadata account
	info.Name = "Unknown"
	info.Symbol = "UNKNOWN"

	return info, nil
}

// GetTokenSupply fetches the total supply of a token
func GetTokenSupply(ctx context.Context, client *rpc.Client, mintAddress string) (uint64, error) {
	mint, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		return 0, fmt.Errorf("invalid mint address: %w", err)
	}

	supply, err := client.GetTokenSupply(ctx, mint, rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get token supply: %w", err)
	}

	// Parse amount string to uint64
	var amount uint64
	fmt.Sscanf(supply.Value.Amount, "%d", &amount)
	return amount, nil
}

// GetTokenHolders fetches all token account holders
// Note: This is expensive on mainnet and may hit rate limits
func GetTokenHolders(ctx context.Context, client *rpc.Client, mintAddress string) ([]TokenAccountInfo, error) {
	mint, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid mint address: %w", err)
	}

	// Get program accounts for the token program filtered by mint
	// This is expensive and may not work well with free RPC
	filters := []rpc.RPCFilter{
		{
			Memcmp: &rpc.RPCFilterMemcmp{
				Offset: 0,
				Bytes:  solana.Base58(mint.Bytes()),
			},
		},
		{
			DataSize: 165, // SPL token account size
		},
	}

	accounts, err := client.GetProgramAccountsWithOpts(
		ctx,
		solana.TokenProgramID,
		&rpc.GetProgramAccountsOpts{
			Filters: filters,
			Encoding: solana.EncodingBase64,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token accounts: %w", err)
	}

	holders := make([]TokenAccountInfo, 0)
	for _, account := range accounts {
		if account.Account.Data == nil {
			continue
		}

		data := account.Account.Data.GetBinary()
		if len(data) < 72 {
			continue
		}

		// Parse token account data
		// Offset 0-32: mint
		// Offset 32-64: owner
		// Offset 64-72: amount (u64)
		
		var owner solana.PublicKey
		copy(owner[:], data[32:64])

		amount := uint64(data[64]) | 
			uint64(data[65])<<8 |
			uint64(data[66])<<16 |
			uint64(data[67])<<24 |
			uint64(data[68])<<32 |
			uint64(data[69])<<40 |
			uint64(data[70])<<48 |
			uint64(data[71])<<56

		if amount > 0 {
			holders = append(holders, TokenAccountInfo{
				Owner:   owner,
				Amount:  amount,
				Address: account.Pubkey,
			})
		}
	}

	return holders, nil
}

// AnalyzeHolderDistribution calculates holder concentration metrics
func AnalyzeHolderDistribution(holders []TokenAccountInfo) (holderCount int, topHolderPct float64, top10Pct float64) {
	if len(holders) == 0 {
		return 0, 0, 0
	}

	// Calculate total supply
	var totalSupply uint64
	for _, holder := range holders {
		totalSupply += holder.Amount
	}

	if totalSupply == 0 {
		return len(holders), 0, 0
	}

	// Sort by amount (descending) - simple bubble sort for small data
	sorted := make([]TokenAccountInfo, len(holders))
	copy(sorted, holders)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Amount < sorted[j+1].Amount {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// Top holder percentage
	if len(sorted) > 0 {
		topHolderPct = float64(sorted[0].Amount) / float64(totalSupply) * 100
	}

	// Top 10 holders percentage
	var top10Supply uint64
	for i := 0; i < 10 && i < len(sorted); i++ {
		top10Supply += sorted[i].Amount
	}
	top10Pct = float64(top10Supply) / float64(totalSupply) * 100

	return len(holders), topHolderPct, top10Pct
}
