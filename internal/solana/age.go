package solana

import (
	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// GetTokenAge fetches the age of a token by finding its first transaction
func GetTokenAge(ctx context.Context, client *rpc.Client, mintAddress string) (time.Time, error) {
	mint, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid mint address: %w", err)
	}

	// Get signatures for this account (oldest first)
	sigs, err := client.GetSignaturesForAddress(ctx, mint)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get signatures: %w", err)
	}

	if len(sigs) == 0 {
		// No transactions found, likely very new or error
		return time.Now(), nil
	}

	// Get the oldest signature (last in the list)
	oldestSig := sigs[len(sigs)-1]
	
	if oldestSig.BlockTime == nil {
		return time.Now(), nil
	}

	// BlockTime is Unix timestamp
	createdAt := time.Unix(int64(*oldestSig.BlockTime), 0)
	return createdAt, nil
}

// GetTokenAgeSeconds returns age in seconds
func GetTokenAgeSeconds(ctx context.Context, client *rpc.Client, mintAddress string) (int64, error) {
	createdAt, err := GetTokenAge(ctx, client, mintAddress)
	if err != nil {
		return 0, err
	}

	age := time.Since(createdAt).Seconds()
	return int64(age), nil
}

// IsTokenTooOld checks if token exceeds max age
func IsTokenTooOld(ctx context.Context, client *rpc.Client, mintAddress string, maxAgeSeconds int64) (bool, int64, error) {
	ageSeconds, err := GetTokenAgeSeconds(ctx, client, mintAddress)
	if err != nil {
		return false, 0, err
	}

	return ageSeconds > maxAgeSeconds, ageSeconds, nil
}
