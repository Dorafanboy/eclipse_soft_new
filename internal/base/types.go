package base

import (
	"eclipse/configs"
	"eclipse/internal/token"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"math"
	"math/rand"
	"time"
)

type SwapConfig struct {
	ProgramID     solana.PublicKey
	PoolAddress   solana.PublicKey
	StateAddress  solana.PublicKey
	VaultA        solana.PublicKey
	VaultB        solana.PublicKey
	PoolAuthority solana.PublicKey
	FeeAccount    solana.PublicKey
	FeeState      solana.PublicKey
	OracleAddress solana.PublicKey
	SwapData      []byte
	NeedsWSol     bool
}

func GetRandomPrecision(minPrecision, maxPrecision int) int {
	return minPrecision + rand.Intn(maxPrecision-minPrecision+1)
}

func RoundFloat(num float64, precision int) float64 {
	multiplier := math.Pow10(precision)
	return math.Round(num*multiplier) / multiplier
}

func GetRandomTokenPair(tokens []configs.Token) (configs.Token, configs.Token, string, error) {
	if len(tokens) < 2 {
		return configs.Token{}, configs.Token{}, "", fmt.Errorf("insufficient tokens: need at least 2, got %d", len(tokens))
	}

	rand.Seed(time.Now().UnixNano())
	firstIndex := rand.Intn(len(tokens))
	firstToken := tokens[firstIndex]

	remainingTokens := make([]configs.Token, 0)
	for i, token := range tokens {
		if i != firstIndex {
			remainingTokens = append(remainingTokens, token)
		}
	}

	secondToken := remainingTokens[rand.Intn(len(remainingTokens))]

	pairType := token.GetPairType(firstToken)

	return firstToken, secondToken, pairType, nil
}
