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
	
	maxAttempts := 15
	for i := 0; i < maxAttempts; i++ {
		response, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		if err != nil {
			if i < maxAttempts-1 {
				logger.Info("Attempt %d: Waiting for confirmation... (%v)", i+1, err)
				time.Sleep(time.Second * 3)
				continue
			}
			return sig, fmt.Errorf("transaction failed: %v", err)
		}

		if response != nil {
			if response.Meta.Err != nil {
				return sig, fmt.Errorf("transaction failed with error: %v", response.Meta.Err)
			}
			logger.Success("Transaction sent succesfully: %s%s", constants.EclipseScan, sig)
			return sig, nil
		}

		time.Sleep(time.Second * 3)
	}

	return sig, fmt.Errorf("transaction not confirmed after %d attempts", maxAttempts)
}
