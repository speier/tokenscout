package solana

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type Client struct {
	rpc    *rpc.Client
	wallet *Wallet
}

func NewClient(rpcURL string, wallet *Wallet) *Client {
	return &Client{
		rpc:    rpc.New(rpcURL),
		wallet: wallet,
	}
}

func (c *Client) GetBalance(ctx context.Context) (uint64, error) {
	balance, err := c.rpc.GetBalance(ctx, c.wallet.PublicKey, rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance.Value, nil
}

func (c *Client) GetBalanceForAddress(ctx context.Context, address solana.PublicKey) (uint64, error) {
	balance, err := c.rpc.GetBalance(ctx, address, rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return balance.Value, nil
}

func (c *Client) GetRecentBlockhash(ctx context.Context) (solana.Hash, error) {
	recent, err := c.rpc.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Hash{}, fmt.Errorf("failed to get recent blockhash: %w", err)
	}
	return recent.Value.Blockhash, nil
}

func (c *Client) SendTransaction(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	sig, err := c.rpc.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight: false,
		PreflightCommitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %w", err)
	}
	return sig, nil
}

func (c *Client) ConfirmTransaction(ctx context.Context, sig solana.Signature) error {
	_, err := c.rpc.GetSignatureStatuses(ctx, true, sig)
	if err != nil {
		return fmt.Errorf("failed to confirm transaction: %w", err)
	}
	return nil
}

func (c *Client) GetWallet() *Wallet {
	return c.wallet
}
