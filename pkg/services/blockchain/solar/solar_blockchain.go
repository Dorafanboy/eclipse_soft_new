package solar

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"encoding/base64"
	"fmt"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"time"
)

func ExecuteSwapFromInstructions(ctx context.Context, client *rpc.Client, encodedTx string, feePayer solana.PrivateKey) (solana.Signature, error) {
	txBytes, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to decode base64: %v", err)
	}

	originalTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return solana.Signature{}, fmt.Errorf("failed to decode transaction: %v", err)
	}

	instructions := make([]solana.Instruction, 0)

	for i, inst := range originalTx.Message.Instructions {
		accounts := make(solana.AccountMetaSlice, 0)
		programID := originalTx.Message.AccountKeys[inst.ProgramIDIndex]

		if i == 0 && programID.Equals(token.ATA_PROGRAM_ID) {
			accounts = append(accounts,
				solana.NewAccountMeta(originalTx.Message.AccountKeys[inst.Accounts[0]], true, true),
				solana.NewAccountMeta(originalTx.Message.AccountKeys[inst.Accounts[1]], false, true),
				solana.NewAccountMeta(originalTx.Message.AccountKeys[inst.Accounts[2]], true, false),
				solana.NewAccountMeta(originalTx.Message.AccountKeys[inst.Accounts[3]], false, false),
				solana.NewAccountMeta(originalTx.Message.AccountKeys[inst.Accounts[4]], false, false),
				solana.NewAccountMeta(originalTx.Message.AccountKeys[inst.Accounts[5]], false, false),
			)
		} else {
			for _, accIdx := range inst.Accounts {
				acc := originalTx.Message.AccountKeys[accIdx]

				accIdxInt := int(accIdx)
				isWritable := false
				isSigner := false

				if accIdxInt < int(originalTx.Message.Header.NumRequiredSignatures-originalTx.Message.Header.NumReadonlySignedAccounts) {
					isWritable = true
				} else if accIdxInt >= int(originalTx.Message.Header.NumRequiredSignatures) &&
					accIdxInt < len(originalTx.Message.AccountKeys)-int(originalTx.Message.Header.NumReadonlyUnsignedAccounts) {
					isWritable = true
				}

				if accIdxInt < int(originalTx.Message.Header.NumRequiredSignatures) {
					isSigner = true
				}

				accounts = append(accounts, solana.NewAccountMeta(acc, isSigner, isWritable))
			}
		}

		newInstruction := solana.NewInstruction(
			programID,
			accounts,
			inst.Data,
		)

		instructions = append(instructions, newInstruction)
	}

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error getting latest blockhash: %v", err)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(feePayer.PublicKey()),
	)
	if err != nil {
		logger.Error("error creating transaction: %v", err)
	}

	tx.Message.SetVersion(0)

	for i := range tx.Message.AccountKeys {
		if tx.Message.AccountKeys[i].Equals(feePayer.PublicKey()) {
			tx.Message.Header.NumRequiredSignatures = 1
			tx.Message.Header.NumReadonlySignedAccounts = 0
			break
		}
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(feePayer.PublicKey()) {
			return &feePayer
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error signing transaction: %v", err)
	}

	retrie := uint(5)
	sig, err := client.SendTransactionWithOpts(ctx, tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
			MaxRetries:          &retrie,
			MinContextSlot:      nil,
		},
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error sending transaction: %v", err)
	}

	time.Sleep(time.Second * 1)

	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		response, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentFinalized,
		})
		if err != nil {
			continue
		}

		if response != nil {
			fmt.Println("Transaction confirmed!")
			break
		}

		if i == maxAttempts-1 {
			fmt.Println("Transaction confirmation timeout")
			break
		}

		time.Sleep(time.Second * 2)
	}

	logger.Success("Transaction sent succesfully: %s%s", constants.EclipseScan, sig)
	return sig, nil
}
