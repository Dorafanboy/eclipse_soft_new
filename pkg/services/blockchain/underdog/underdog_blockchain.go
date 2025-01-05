package underdog

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/logger"
	"encoding/base64"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"time"
)

func SendSolanaTransaction(ctx context.Context, client *rpc.Client, encodedTx string, userPrivateKey solana.PrivateKey) (solana.Signature, error) {
	txBytes, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to decode base64: %v", err)
	}

	tx, err := solana.TransactionFromBytes(txBytes)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to decode transaction: %v", err)
	}

	messageBytes, _ := tx.Message.MarshalBinary()
	newSignature, _ := userPrivateKey.Sign(messageBytes)
	tx.Signatures[0] = newSignature

	retries := uint(5)
	sig, err := client.SendTransactionWithOpts(ctx, tx,
		rpc.TransactionOpts{
			SkipPreflight:       true,
			PreflightCommitment: rpc.CommitmentFinalized,
			MaxRetries:          &retries,
		},
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error sending transaction: %v", err)
	}

	logger.Success("Transaction sent succesfully: %s%s", constants.EclipseScan, sig)

	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(time.Second)
		status, err := client.GetSignatureStatuses(ctx, true, sig)
		if err != nil {
			logger.Error("Error checking status: %v\n", err)
			continue
		}
		if status.Value[0] != nil {
			if status.Value[0].Err != nil {
				return solana.Signature{}, fmt.Errorf("transaction failed: %v", status.Value[0].Err)
			}
			return sig, nil
		}
	}

	return solana.Signature{}, fmt.Errorf("transaction confirmation timeout")
}
