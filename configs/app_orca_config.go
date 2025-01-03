package configs

import (
	"fmt"
	"github.com/gagliardetto/solana-go"
)

type OrcaConfig struct {
	Tokens []Token
	Native NativeConfig
	Stable AmountConfig
}

var availableTokensOrca = map[string]Token{
	"ETH": {
		Symbol:   "ETH",
		Address:  solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112"),
		Decimals: 9,
	},
	"SOL": {
		Symbol:   "SOL",
		Address:  solana.MustPublicKeyFromBase58("BeRUj3h7BqkbdfFU7FBNYbodgf8GCHodzKvF9aVjNNfL"),
		Decimals: 9,
	},
	"USDC": {
		Symbol:   "USDC",
		Address:  solana.MustPublicKeyFromBase58("AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE"),
		Decimals: 6,
	},
	"USDT": {
		Symbol:   "USDT",
		Address:  solana.MustPublicKeyFromBase58("CEBP3CqAbW4zdZA57H2wfaSG1QNdzQ72GiQEbQXyW9Tm"),
		Decimals: 6,
	},
}

func NewOrcaConfig() (*OrcaConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error when reading orca_config: %v", err)
	}

	var configuredTokens []Token
	for _, symbol := range cfg.Orca.Tokens {
		if token, exists := availableTokensOrca[symbol]; exists {
			configuredTokens = append(configuredTokens, token)
		}
	}

	return &OrcaConfig{
		Tokens: configuredTokens,
		Native: cfg.Orca.Native,
		Stable: cfg.Orca.Stable,
	}, nil
}
