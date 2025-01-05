package gas_station

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/logger"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"time"
)

func SendTransaction(ctx context.Context, client *rpc.Client, privateKey solana.PrivateKey, serializedTx string) (solana.Signature, error) {
	txBytes, err := base58.Decode(serializedTx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to decode transaction: %v", err)
	}

	decoder := bin.NewBinDecoder(txBytes)
	tx := &solana.Transaction{}

	if err := tx.UnmarshalWithDecoder(decoder); err != nil {
		return solana.Signature{}, fmt.Errorf("failed to unmarshal transaction: %v", err)
	}

	publicKey := privateKey.PublicKey()

	var signerIndex = -1
	for i, key := range tx.Message.AccountKeys {
		if key.Equals(publicKey) {
			signerIndex = i
			break
		}
	}
	if signerIndex == -1 {
		return solana.Signature{}, fmt.Errorf("transaction does not require signature from %s", publicKey.String())
	}

	messageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failted to unmarshal binary %v", err)
	}
	signature, err := privateKey.Sign(messageBytes)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failted to sign %v", err)
	}

	if len(tx.Signatures) <= signerIndex {
		newSigs := make([]solana.Signature, signerIndex+1)
		copy(newSigs, tx.Signatures)
		tx.Signatures = newSigs
	}

	tx.Signatures[signerIndex] = signature

	retries := uint(5)
	opts := rpc.TransactionOpts{
		SkipPreflight:       true,
		PreflightCommitment: rpc.CommitmentProcessed,
		MaxRetries:          &retries,
		MinContextSlot:      nil,
	}

	sig, err := client.SendTransactionWithOpts(
		ctx,
		tx,
		opts,
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to send transaction: %v", err)
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
