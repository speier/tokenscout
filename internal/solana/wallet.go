package solana

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gagliardetto/solana-go"
)

type Wallet struct {
	PrivateKey solana.PrivateKey
	PublicKey  solana.PublicKey
}

func NewWallet() (*Wallet, error) {
	account := solana.NewWallet()
	return &Wallet{
		PrivateKey: account.PrivateKey,
		PublicKey:  account.PublicKey(),
	}, nil
}

func LoadWallet(path string) (*Wallet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	var keyBytes []byte
	if err := json.Unmarshal(data, &keyBytes); err != nil {
		return nil, fmt.Errorf("failed to parse wallet file: %w", err)
	}

	privateKey := solana.PrivateKey(keyBytes)
	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  privateKey.PublicKey(),
	}, nil
}

func (w *Wallet) Save(path string) error {
	keyBytes := []byte(w.PrivateKey)
	data, err := json.MarshalIndent(keyBytes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

func (w *Wallet) Address() string {
	return w.PublicKey.String()
}

func GenerateMnemonic() (string, error) {
	entropy := make([]byte, 32)
	if _, err := rand.Read(entropy); err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}
	
	// For now, just return a note. Full BIP39 mnemonic generation would require additional library
	return "", fmt.Errorf("mnemonic generation not yet implemented - use wallet file instead")
}
