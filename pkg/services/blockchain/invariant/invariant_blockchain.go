package invariant

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/token"
	"eclipse/pkg/services/blockchain/lifinity"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	TOKEN_PROGRAM_ID     = solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	SYSTEM_PROGRAM_ID    = solana.MustPublicKeyFromBase58("11111111111111111111111111111111")
	INVARIANT_PROGRAM_ID = solana.MustPublicKeyFromBase58("iNvTyprs4TX8m6UeUEkeqDFjAL9zRCRWcexK9Sd4WEU")
)

func createCreateAccountInstruction(params token.SwapInstructions, newAccount solana.PublicKey) solana.Instruction {
	data := make([]byte, 52)

	binary.LittleEndian.PutUint32(data[0:4], 0)
	binary.LittleEndian.PutUint64(data[4:12], 19924)
	binary.LittleEndian.PutUint64(data[12:20], 165)

	ownerPubkey := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")
	copy(data[20:52], ownerPubkey.Bytes())

	return solana.NewInstruction(
		SYSTEM_PROGRAM_ID,
		solana.AccountMetaSlice{
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(newAccount, true, true),
		},
		data,
	)
}

func createTransferInstruction(params token.SwapInstructions, newAccount solana.PublicKey, amount uint64) solana.Instruction {
	transferData := make([]byte, 12)
	transferData[0] = 2
	binary.LittleEndian.PutUint64(transferData[4:], amount)

	return solana.NewInstruction(
		SYSTEM_PROGRAM_ID,
		solana.AccountMetaSlice{
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(newAccount, true, true),
		},
		transferData,
	)
}

func createInitializeAccountInstruction(params token.SwapInstructions, newAccount solana.PublicKey) solana.Instruction {
	RENT_PROGRAM_ID := solana.MustPublicKeyFromBase58("SysvarRent111111111111111111111111111111111")

	return solana.NewInstruction(
		TOKEN_PROGRAM_ID,
		solana.AccountMetaSlice{
			solana.NewAccountMeta(newAccount, true, true),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58(WRAPPED_ETH_ADDRESS), false, false),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(RENT_PROGRAM_ID, false, false),
		},
		[]byte{1},
	)
}

func createSwapInstruction(params token.SwapInstructions, tempAccount solana.PublicKey) solana.Instruction {
	mint := "AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE"

	mintKey := solana.MustPublicKeyFromBase58(mint)

	destinationATA, _, err := token.FindAssociatedTokenAddress2022(params.Payer.PublicKey(), mintKey)
	if err != nil {
		log.Fatalf("Error finding associated token address: %v", err)
	}

	data := make([]byte, 34)
	discriminator := []byte{
		0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8,
	}
	copy(data[0:8], discriminator)

	data[8] = 1

	binary.LittleEndian.PutUint64(data[9:17], params.Amount)

	data[17] = 1

	sqrtPriceLimitHex := "0"
	sqrtPriceLimitBytes, _ := hex.DecodeString(sqrtPriceLimitHex)
	copy(data[18:34], sqrtPriceLimitBytes)

	var accountMetas solana.AccountMetaSlice

	if params.IsETH {
		accountMetas = solana.AccountMetaSlice{
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("J1GyvGaY3XPNGFQkMNx7C5aT48duYRrAJE75Evk3neas"), false, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("FdEcxaJ9cDW5Y3AZ79eDtzdvK7uaxSEfn5vPb47ew5yg"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("4qw2sfWzrkhVWEHfXJsG7g7h6hyGuwsaL3Cit8u5XZsj"), true, false),
			solana.NewAccountMeta(params.SecondToken, false, false),
			solana.NewAccountMeta(params.FirstToken, false, false),
			solana.NewAccountMeta(destinationATA, true, false),
			solana.NewAccountMeta(tempAccount, true, true),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("EUC57x63cdmsCRFDPtZcd5Ko8T4qmA8RRcvun96N96Jv"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("6pgHDYgEzzYT8e1em6r5kchp3PJ8E5gfR6chhSgT9o1P"), true, false),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("D4P9HJYPczLFHvxBgpLKooy7eWczci8pr4x9Zu7iYCVN"), false, false),
			solana.NewAccountMeta(lifinity.TOKEN_2022_PROGRAM_ID, false, false),
			solana.NewAccountMeta(TOKEN_PROGRAM_ID, false, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("EwG2ZsoZvXW5Mp7Gc2VXrSkEehTH9B7EntCmR8fcnQtQ"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("95vhcMhq1nS4pyYwV5dd2GiLx7Vux1VUm74o68QGbTuw"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("B3QN7GUiZQbMb5sxQkEYd1fjuxSRHQQ8aUY8HWjLSaaR"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("zscN6fjzXzj1fmHCU6ey5w51SxXyZ7qWdb4RiABNoDM"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("FRhLPgegA1sbr4zfcD7xSbRRVsGu7QstnUCeCV3GCJDq"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("FBJwWWR2eDfJTQpR9FYqYjQcsx8fCXdpJ5gXQSpQvy8z"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("4zkqTKeoCVBaHn8NZXJmUSic7LWQnooLXBuRmbcSgbrC"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("7i429eLMEdLxYaTJRjEffUr1UdRFr5zrCX6vNeZFKtcR"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("6AqMpWfWxUdoxqbiwAwGG2LjJyEcsq9dHH2Qi4Bxfcot"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("J87LRSm5dAqFVm5G4D91d6HRvTVx8hQR1zP6Lc7HYtpZ"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("4KqauEUpfQpDQF2MJigT6ZZPoTqmgoMf7Yo7PKhy2BZh"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("9sRgpBm8idu7R8u8XnFGCz9enhvrEegZzkQBZkTG6eLY"), true, false),
		}
	} else {
		accountMetas = solana.AccountMetaSlice{
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("J1GyvGaY3XPNGFQkMNx7C5aT48duYRrAJE75Evk3neas"), false, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("HRgVv1pyBLXdsAddq4ubSqo8xdQWRrYbvmXqEDtectce"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("5DT9Vdod6JgqJhdbWSDXt6CaLVAHZB5DjxDjPDQ5rqPv"), true, false),
			solana.NewAccountMeta(params.FirstToken, false, false),
			solana.NewAccountMeta(params.SecondToken, false, false),
			solana.NewAccountMeta(destinationATA, true, false),
			solana.NewAccountMeta(tempAccount, true, true),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("4AY5LB93QjtsdZ8Ln5M7HJrLkjmxR8xAt6qCwPmTeL8p"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("44wJFmJbZWDbENLAn7PHE29ULZzC8EiBettkT48ePPT7"), true, false),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("D4P9HJYPczLFHvxBgpLKooy7eWczci8pr4x9Zu7iYCVN"), false, false),
			solana.NewAccountMeta(lifinity.TOKEN_2022_PROGRAM_ID, false, false),
			solana.NewAccountMeta(TOKEN_PROGRAM_ID, false, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("Dhn3qtRg64qy3zS92Py9km1H1K45C5FyHtfomfF6TLvH"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("8Hd1P8TE3g9daaa1HjJXPDg315FhaGaeptuNvUk3Dw9f"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("5A1JsrPPPpejc3YNfWjeySCEEzG4xJcFS8w2XFdc3LY4"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("HcceiFnd7MV6ZpMVN4sYxeu8xXpFvWEGaRpagYYKEH8K"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("4Ae7GfgJtJGiFF1wosm7SPGifHvsgwdav2mYbAPvkSuH"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("77stihPdVQEMyTJZtpQSyDrmpNxZAHfY8pRE7hoLKk7c"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("44Qp49CisK1QHb78ws2ZenaUu3eH4tRxYfiePZkYd7aT"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("5AsXGNnrQjFoTwqS53kt6Sw9WDtDxE4b3tGcKLZPt19G"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("JBhWCkxJAu6WdPBzxufVSmXWPUVULK3tAvWSuvQa9AfB"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("DkiQZDoEmWMgVbVipzkz8KSsVrikeAGFXBd4Kyjagzh5"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("4b45YauzzMv8W8UruSVBe42yTeap7A7Lidy7nBNJVsWv"), true, false),
			solana.NewAccountMeta(solana.MustPublicKeyFromBase58("G1oBuHHzf9wtBvgA588RuBbuLjZ7Vwgh3E8jfh29Y33B"), true, false),
		}
	}

	return solana.NewInstruction(
		INVARIANT_PROGRAM_ID,
		accountMetas,
		data,
	)
}

func createCloseAccountInstruction(params token.SwapInstructions, tempAccount solana.PublicKey) solana.Instruction {
	return solana.NewInstruction(
		TOKEN_PROGRAM_ID,
		solana.AccountMetaSlice{
			solana.NewAccountMeta(tempAccount, true, true),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
		},
		[]byte{9},
	)
}

func CreateFullSwapInstructions(params token.SwapInstructions) ([]solana.Instruction, *solana.PrivateKey, error) {
	newAccountKeypair := solana.NewWallet().PrivateKey
	newAccountPubkey := newAccountKeypair.PublicKey()

	createAccountIx := createCreateAccountInstruction(params, newAccountKeypair.PublicKey())

	instructions := []solana.Instruction{
		createAccountIx,
	}

	if params.IsETH {
		instructions = append(instructions, createTransferInstruction(params, newAccountPubkey, params.Amount))
	}

	instructions = append(instructions,
		createInitializeAccountInstruction(params, newAccountPubkey),
		createSwapInstruction(params, newAccountPubkey),
		createCloseAccountInstruction(params, newAccountPubkey),
	)

	return instructions, &newAccountKeypair, nil
}

func InvariantSendTx(ctx context.Context, client *rpc.Client, instructions []solana.Instruction, feePayer solana.PrivateKey, newAccountKeypair *solana.PrivateKey) (solana.Signature, error) {
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
		return solana.Signature{}, fmt.Errorf("error creating transaction: %v", err)
	}

	tx.Message.SetVersion(0)

	for i := range tx.Message.AccountKeys {
		if tx.Message.AccountKeys[i].Equals(feePayer.PublicKey()) ||
			tx.Message.AccountKeys[i].Equals(newAccountKeypair.PublicKey()) {
			tx.Message.Header.NumRequiredSignatures = 2
			tx.Message.Header.NumReadonlySignedAccounts = 0
			break
		}
	}

	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(feePayer.PublicKey()) {
			return &feePayer
		}
		if key.Equals(newAccountKeypair.PublicKey()) {
			return newAccountKeypair
		}
		return nil
	})
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error signing transaction: %v", err)
	}

	retrie := uint(10)
	sig, err := client.SendTransactionWithOpts(ctx, tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentConfirmed,
			MaxRetries:          &retrie,
			MinContextSlot:      nil,
		},
	)
	if err != nil {
		return solana.Signature{}, fmt.Errorf("error sending transaction: %v", err)
	}

	time.Sleep(time.Second * 2)

	maxAttempts := 15
	for i := 0; i < maxAttempts; i++ {
		response, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		if err != nil {
			if i < maxAttempts-1 {
				log.Printf("Attempt %d: Waiting for confirmation... (%v)", i+1, err)
				time.Sleep(time.Second * 3)
				continue
			}
			return sig, fmt.Errorf("transaction failed: %v", err)
		}

		if response != nil {
			if response.Meta.Err != nil {
				return sig, fmt.Errorf("transaction failed with error: %v", response.Meta.Err)
			}
			log.Printf("Transaction sent succesfully: %s%s", constants.EclipseScan, sig)
			return sig, nil
		}

		time.Sleep(time.Second * 3)
	}

	return sig, fmt.Errorf("transaction not confirmed after %d attempts", maxAttempts)
}
