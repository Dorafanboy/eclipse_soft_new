package orca

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/logger"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"time"
)

type SpecialAccount struct {
	IsSigner   bool
	IsWritable bool
	IsFeePayer bool
}

func SimulateAndSendTransaction(ctx context.Context, client *rpc.Client, instructions *SwapInstructions, privateKey solana.PrivateKey) (solana.Signature, error) {
	var solanaInstructions []solana.Instruction

	var secondSignerPrivateKey solana.PrivateKey
	if len(instructions.Data.Signers) > 0 {
		secondSignerPrivateKey = instructions.Data.Signers[0]
	}

	for _, inst := range instructions.Data.Instructions {
		programID := solana.PublicKeyFromBytes(inst.ProgramID)
		accounts := make(solana.AccountMetaSlice, len(inst.Accounts))

		for i, acc := range inst.Accounts {
			pubKey := solana.PublicKeyFromBytes(acc.Pubkey)
			accounts[i] = &solana.AccountMeta{
				PublicKey:  pubKey,
				IsSigner:   acc.IsSigner,
				IsWritable: acc.IsWritable,
			}
		}

		instruction := solana.NewInstruction(
			programID,
			accounts,
			inst.Data,
		)
		solanaInstructions = append(solanaInstructions, instruction)
	}

	recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error getting latest blockhash: %v", err)
	}

	tx, err := solana.NewTransaction(
		solanaInstructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(privateKey.PublicKey()),
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error creating transaction: %v", err)
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(privateKey.PublicKey()) {
			return &privateKey
		}
		if len(secondSignerPrivateKey) > 0 && key.Equals(secondSignerPrivateKey.PublicKey()) {
			return &secondSignerPrivateKey
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
		return solana.Signature{}, fmt.Errorf("transaction execution failed: %v", err)
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
