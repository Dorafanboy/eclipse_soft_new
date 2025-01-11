package balance

import (
	"context"
	"eclipse/internal/token"
	"fmt"
	"strconv"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	USDC = solana.MustPublicKeyFromBase58("AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE")
)

func GetUSDCBalance(ctx context.Context, rpcClient *rpc.Client, publicKey solana.PublicKey) (uint64, error) {
	tokenAccount, _, err := token.FindAssociatedTokenAddress2022(
		publicKey,
		USDC,
	)
	if err != nil {
		return 0, fmt.Errorf("error getting token account: %v", err)
	}

	balance, err := rpcClient.GetTokenAccountBalance(
		ctx,
		tokenAccount,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return 0, fmt.Errorf("error getting balance: %v", err)
	}

	amount, err := strconv.ParseUint(balance.Value.Amount, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing balance: %v", err)
	}

	return amount, nil
}
