package engine

import (
	"encoding/hex"

	"github.com/gagliardetto/solana-go"
	"github.com/speier/tokenscout/internal/logger"
)

// DEX program addresses
var (
	RaydiumAMMV4 = solana.MustPublicKeyFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8")
	OrcaWhirlpool = solana.MustPublicKeyFromBase58("9W959DqEETiGZocYWCQPaJ6sBmUzgfxXfqGeTEdp3aQP")
	WrappedSOL = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
)

// InstructionParser defines the interface for DEX-specific parsers
type InstructionParser interface {
	// CanParse checks if this parser can handle the instruction
	CanParse(programID solana.PublicKey, accounts []solana.PublicKey, data []byte) bool
	
	// ParseTokenMint extracts the new token mint from the instruction
	ParseTokenMint(accounts []solana.PublicKey, data []byte) (string, bool)
	
	// Name returns the parser name for logging
	Name() string
}

// RaydiumParser handles Raydium AMM V4 pool initialization
type RaydiumParser struct{}

func (p *RaydiumParser) Name() string {
	return "Raydium AMM V4"
}

func (p *RaydiumParser) CanParse(programID solana.PublicKey, accounts []solana.PublicKey, data []byte) bool {
	// Check if it's Raydium program
	if !programID.Equals(RaydiumAMMV4) {
		return false
	}
	
	// Raydium initialize instructions have many accounts (typically 17+)
	// and specific instruction discriminators
	if len(accounts) < 15 {
		return false
	}
	
	// Check instruction discriminator (first 8 bytes)
	// Raydium uses multiple initialize variants
	if len(data) < 8 {
		return false
	}
	
	discriminator := hex.EncodeToString(data[:8])
	
	// Known Raydium initialize discriminators:
	// - initialize: 0xafaf6d1f0d989bed (legacy)
	// - initialize2: 0x95b7328b94c29e09 (current)
	// We'll accept any instruction with sufficient accounts for now
	// and validate by checking for token mints
	
	logger.Debug().
		Str("discriminator", discriminator).
		Int("accounts", len(accounts)).
		Msg("Checking Raydium instruction")
	
	return true
}

func (p *RaydiumParser) ParseTokenMint(accounts []solana.PublicKey, data []byte) (string, bool) {
	// Raydium AMM V4 initialize/initialize2 account layout:
	// [0] Token program
	// [1] System program
	// [2] Rent sysvar
	// [3] AMM ID
	// [4] AMM authority
	// [5] AMM open orders
	// [6] LP mint address
	// [7] Coin mint (token A)
	// [8] PC mint (token B, usually SOL)
	// [9] Pool coin token account
	// [10] Pool pc token account
	// ... more accounts
	
	// Require at least 9 accounts for safe access
	if len(accounts) <= 8 {
		logger.Debug().
			Int("accounts", len(accounts)).
			Msg("Raydium: Not enough accounts")
		return "", false
	}
	
	tokenA := accounts[7]
	tokenB := accounts[8]
	
	logger.Debug().
		Str("token_a", tokenA.String()).
		Str("token_b", tokenB.String()).
		Msg("Raydium: Found token pair")
	
	// Return the non-SOL token
	if !tokenA.Equals(WrappedSOL) {
		return tokenA.String(), true
	}
	if !tokenB.Equals(WrappedSOL) {
		return tokenB.String(), true
	}
	
	// If both are not SOL, return the first one
	// (this might be a token-token pair)
	return tokenA.String(), true
}

// OrcaParser handles Orca Whirlpool pool initialization
type OrcaParser struct{}

func (p *OrcaParser) Name() string {
	return "Orca Whirlpool"
}

func (p *OrcaParser) CanParse(programID solana.PublicKey, accounts []solana.PublicKey, data []byte) bool {
	// Check if it's Orca program
	if !programID.Equals(OrcaWhirlpool) {
		return false
	}
	
	// Orca initialize instructions have specific account count
	if len(accounts) < 8 {
		return false
	}
	
	if len(data) < 8 {
		return false
	}
	
	discriminator := hex.EncodeToString(data[:8])
	
	logger.Debug().
		Str("discriminator", discriminator).
		Int("accounts", len(accounts)).
		Msg("Checking Orca instruction")
	
	return true
}

func (p *OrcaParser) ParseTokenMint(accounts []solana.PublicKey, data []byte) (string, bool) {
	// Orca Whirlpool initializePool account layout:
	// [0] WhirlpoolsConfig
	// [1] Token mint A
	// [2] Token mint B  
	// [3] Funder
	// [4] Whirlpool PDA
	// [5] Token vault A
	// [6] Token vault B
	// [7] Fee tier
	// ... more accounts
	
	// Require at least 3 accounts for safe access
	if len(accounts) <= 2 {
		logger.Debug().
			Int("accounts", len(accounts)).
			Msg("Orca: Not enough accounts")
		return "", false
	}
	
	tokenA := accounts[1]
	tokenB := accounts[2]
	
	logger.Debug().
		Str("token_a", tokenA.String()).
		Str("token_b", tokenB.String()).
		Msg("Orca: Found token pair")
	
	// Return the non-SOL token
	if !tokenA.Equals(WrappedSOL) {
		return tokenA.String(), true
	}
	if !tokenB.Equals(WrappedSOL) {
		return tokenB.String(), true
	}
	
	// If both are not SOL, return the first one
	return tokenA.String(), true
}

// ParsersRegistry holds all available parsers
type ParsersRegistry struct {
	parsers []InstructionParser
}

func NewParsersRegistry() *ParsersRegistry {
	return &ParsersRegistry{
		parsers: []InstructionParser{
			&RaydiumParser{},
			&OrcaParser{},
		},
	}
}

func (r *ParsersRegistry) ParseInstruction(programID solana.PublicKey, accounts []solana.PublicKey, data []byte) (string, string, bool) {
	for _, parser := range r.parsers {
		if parser.CanParse(programID, accounts, data) {
			if mint, ok := parser.ParseTokenMint(accounts, data); ok {
				return mint, parser.Name(), true
			}
		}
	}
	return "", "", false
}
