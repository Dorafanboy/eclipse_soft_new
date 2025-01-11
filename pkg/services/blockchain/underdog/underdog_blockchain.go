package underdog

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/logger"
	"encoding/base64"
	"encoding/binary"
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

func EstimateTransactionFee(ctx context.Context, client *rpc.Client, encodedTx string) (uint64, error) {
	txBytes, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64: %v", err)
	}

	tx, err := solana.TransactionFromBytes(txBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to decode transaction: %v", err)
	}

	response, err := client.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent blockhash: %v", err)
	}

	feeRate := response.Value.FeeCalculator.LamportsPerSignature
	totalFee := feeRate * uint64(len(tx.Signatures))

	return totalFee, nil
}

func GetTransactionAmount(encodedTx string) (uint64, error) {
	txBytes, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return 0, fmt.Errorf("failed to decode base64: %v", err)
	}

	tx, err := solana.TransactionFromBytes(txBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to decode transaction: %v", err)
	}

	if len(tx.Message.Instructions) > 0 {
		instruction := tx.Message.Instructions[0]
		if len(instruction.Data) >= 8 {
			amount := binary.LittleEndian.Uint64(instruction.Data[:8])
			return amount, nil
		}
	}

	return 0, fmt.Errorf("could not find amount in transaction")
}
