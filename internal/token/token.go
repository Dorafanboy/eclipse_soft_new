package token

import (
	"eclipse/configs"
	"github.com/gagliardetto/solana-go"
)

type SwapInstructions struct {
	Payer         solana.PrivateKey
	FirstToken    solana.PublicKey
	SecondToken   solana.PublicKey
	Amount        uint64
	IsETH         bool
	TokenSymbol   string
	TokenDecimals int8
}

var TOKEN_2022_PROGRAM_ID = solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
var ATA_PROGRAM_ID = solana.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")

func FindAssociatedTokenAddress2022(wallet, mint solana.PublicKey) (solana.PublicKey, uint8, error) {
	seeds := [][]byte{
		wallet.Bytes(),
		TOKEN_2022_PROGRAM_ID.Bytes(),
		mint.Bytes(),
	}
	return solana.FindProgramAddress(seeds, ATA_PROGRAM_ID)
}

func GetPairType(first configs.Token) string {
	if first.Symbol == "USDC" || first.Symbol == "USDT" {
		return "stable"
	}

	if first.Symbol == "ETH" {
		return "ETH"
	}

	if first.Symbol == "SOL" {
		return "SOL"
	}

	return "unknown"
}
