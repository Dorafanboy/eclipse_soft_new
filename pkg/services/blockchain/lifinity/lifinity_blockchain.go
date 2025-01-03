package lifinity

import (
	"context"
	"eclipse/constants"
	"eclipse/internal/base"
	"eclipse/internal/token"
	"encoding/binary"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"math"
	"math/rand"
	"time"
)

var (
	TOKEN_2022_PROGRAM_ID = solana.MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")
	SYSTEM_PROGRAM_ID     = solana.MustPublicKeyFromBase58("11111111111111111111111111111111")
	ATA_PROGRAM_ID        = solana.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")
	COMPUTE_BUDGET_ID     = solana.MustPublicKeyFromBase58("ComputeBudget111111111111111111111111111111")
	LIFINITY_PROGRAM_ID   = solana.MustPublicKeyFromBase58("4UsSbJQZJTfZDFrgvcPBRCSg5BbcQE6dobnriCafzj12")
	MINT_DATA             = solana.MustPublicKeyFromBase58("9pan9bMn5HatX4EJdBwg9VgCa7Uz5HL8N1m5D3NdXejP")
	MIN_COMPUTE_UNITS     = 131000
	MAX_COMPUTE_UNITS     = 138000
)

func createSwapConfig(baseConfig base.SwapConfig, swapData []byte, needsWSol bool) base.SwapConfig {
	config := baseConfig
	config.SwapData = swapData
	config.NeedsWSol = needsWSol
	return config
}

func createCloseAccountInstruction(params token.SwapInstructions) (solana.Instruction, error) {
	sourceATA, _, err := token.FindAssociatedTokenAddress2022(params.Payer.PublicKey(), MINT_DATA)
	if err != nil {
		return nil, err
	}

	return solana.NewInstruction(
		TOKEN_2022_PROGRAM_ID,
		solana.AccountMetaSlice{
			solana.NewAccountMeta(sourceATA, true, false),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
		},
		[]byte{9},
	), nil
}

func createWSolInstructions(params token.SwapInstructions, sourceATA solana.PublicKey) []solana.Instruction {
	transferAmount := uint64(100000)
	transferData := make([]byte, 12)
	transferData[0] = 2
	binary.LittleEndian.PutUint64(transferData[4:], transferAmount)

	return []solana.Instruction{
		solana.NewInstruction(
			SYSTEM_PROGRAM_ID,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
				solana.NewAccountMeta(sourceATA, false, true),
			},
			transferData,
		),
		solana.NewInstruction(
			TOKEN_2022_PROGRAM_ID,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(sourceATA, false, true),
				solana.NewAccountMeta(params.FirstToken, false, false),
				solana.NewAccountMeta(params.Payer.PublicKey(), false, false),
			},
			[]byte{0x11},
		),
	}
}

func calculateRequiredComputeUnits() uint32 {
	rand.Seed(time.Now().UnixNano())
	units := rand.Intn(MAX_COMPUTE_UNITS-MIN_COMPUTE_UNITS+1) + MIN_COMPUTE_UNITS

	return uint32(units)
}

func createCommonInstructions(ctx context.Context, params token.SwapInstructions, config base.SwapConfig) ([]solana.Instruction, error) {
	instructions := make([]solana.Instruction, 0)

	computeUnits := calculateRequiredComputeUnits()
	computeUnitLimitIx := solana.NewInstruction(
		COMPUTE_BUDGET_ID,
		solana.AccountMetaSlice{},
		[]byte{
			0x02,
			byte(computeUnits & 0xff),
			byte((computeUnits >> 8) & 0xff),
			byte((computeUnits >> 16) & 0xff),
			byte((computeUnits >> 24) & 0xff),
		},
	)

	instructions = append(instructions, computeUnitLimitIx)

	sourceATA, _, err := token.FindAssociatedTokenAddress2022(params.Payer.PublicKey(), MINT_DATA)

	destinationATAFirst, _, err := token.FindAssociatedTokenAddress2022(params.Payer.PublicKey(), params.FirstToken)
	if err != nil {
		return nil, fmt.Errorf("error finding destination ATA: %v", err)
	}

	destinationATASecond, _, err := token.FindAssociatedTokenAddress2022(params.Payer.PublicKey(), params.SecondToken)
	if err != nil {
		return nil, fmt.Errorf("error finding destination ATA: %v", err)
	}

	var ataInstruction *solana.GenericInstruction = nil
	if params.IsETH {
		ataInstruction = solana.NewInstruction(
			ATA_PROGRAM_ID,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
				solana.NewAccountMeta(sourceATA, true, false),
				solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
				solana.NewAccountMeta(MINT_DATA, true, false),
				solana.NewAccountMeta(SYSTEM_PROGRAM_ID, false, false),
				solana.NewAccountMeta(TOKEN_2022_PROGRAM_ID, false, false),
			},
			[]byte{},
		)

		instructions = append(instructions, ataInstruction)

		transferData := make([]byte, 12)
		transferData[0] = 2
		binary.LittleEndian.PutUint64(transferData[4:], params.Amount)

		systemTransferIx := solana.NewInstruction(
			SYSTEM_PROGRAM_ID,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
				solana.NewAccountMeta(sourceATA, true, false),
			},
			transferData,
		)
		instructions = append(instructions, systemTransferIx)

		syncNativeIx := solana.NewInstruction(
			TOKEN_2022_PROGRAM_ID,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(sourceATA, true, false),
			},
			[]byte{
				0x11,
			},
		)
		instructions = append(instructions, syncNativeIx)
	} else {
		ataInstruction = solana.NewInstruction(
			ATA_PROGRAM_ID,
			solana.AccountMetaSlice{
				solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
				solana.NewAccountMeta(sourceATA, true, false),
				solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
				solana.NewAccountMeta(MINT_DATA, true, false),
				solana.NewAccountMeta(SYSTEM_PROGRAM_ID, false, false),
				solana.NewAccountMeta(TOKEN_2022_PROGRAM_ID, false, false),
			},
			[]byte{1},
		)

		instructions = append(instructions, ataInstruction)
	}

	//if config.NeedsWSol {
	//	wsolInstructions := createWSolInstructions(params, sourceATA)
	//	instructions = append(instructions, wsolInstructions...)
	//}

	var swapAccounts []*solana.AccountMeta = nil
	if params.IsETH {
		swapAccounts = solana.AccountMetaSlice{
			solana.NewAccountMeta(config.PoolAddress, false, false),
			solana.NewAccountMeta(config.StateAddress, true, false),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(sourceATA, true, false),
			solana.NewAccountMeta(destinationATASecond, true, false),
			solana.NewAccountMeta(config.VaultA, true, false),
			solana.NewAccountMeta(config.VaultB, true, false),
			solana.NewAccountMeta(MINT_DATA, true, false),
			solana.NewAccountMeta(params.SecondToken, true, false),
			solana.NewAccountMeta(config.FeeAccount, true, false),
			solana.NewAccountMeta(config.FeeState, true, false),
			solana.NewAccountMeta(TOKEN_2022_PROGRAM_ID, false, false),
			solana.NewAccountMeta(config.PoolAuthority, false, false),
			solana.NewAccountMeta(config.PoolAuthority, false, false),
			solana.NewAccountMeta(config.OracleAddress, false, false),
		}
	} else {
		swapAccounts = solana.AccountMetaSlice{
			solana.NewAccountMeta(config.PoolAddress, false, false),
			solana.NewAccountMeta(config.StateAddress, true, false),
			solana.NewAccountMeta(params.Payer.PublicKey(), true, true),
			solana.NewAccountMeta(destinationATAFirst, true, false),
			solana.NewAccountMeta(sourceATA, true, false),
			solana.NewAccountMeta(config.VaultB, true, false),
			solana.NewAccountMeta(config.VaultA, true, false),
			solana.NewAccountMeta(params.FirstToken, true, false),
			solana.NewAccountMeta(MINT_DATA, true, false),
			solana.NewAccountMeta(config.FeeAccount, true, false),
			solana.NewAccountMeta(config.FeeState, true, false),
			solana.NewAccountMeta(TOKEN_2022_PROGRAM_ID, false, false),
			solana.NewAccountMeta(config.PoolAuthority, false, false),
			solana.NewAccountMeta(config.PoolAuthority, false, false),
			solana.NewAccountMeta(config.OracleAddress, false, false),
		}
	}

	swapIx := solana.NewInstruction(config.ProgramID, swapAccounts, config.SwapData)
	instructions = append(instructions, swapIx)

	return instructions, nil
}

func CreateFullSwapInstructionsFromEthToUsdc(ctx context.Context, params token.SwapInstructions) ([]solana.Instruction, error) {
	swapData := make([]byte, 24)
	copy(swapData[0:8], []byte{0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8})
	binary.LittleEndian.PutUint64(swapData[8:16], params.Amount)
	binary.LittleEndian.PutUint64(swapData[16:24], 0)

	config := createSwapConfig(ETH_USDC_POOL, swapData, false)

	instructions, err := createCommonInstructions(ctx, params, config)
	if err != nil {
		return nil, err
	}

	closeAccountIx, err := createCloseAccountInstruction(params)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, closeAccountIx)

	return instructions, nil
}

func CreateFullSwapInstructionsFromUsdcToEth(ctx context.Context, params token.SwapInstructions) ([]solana.Instruction, error) {
	swapData := make([]byte, 24)
	copy(swapData[0:8], []byte{0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8})
	binary.LittleEndian.PutUint64(swapData[8:16], params.Amount)
	binary.LittleEndian.PutUint64(swapData[16:24], 0)

	config := createSwapConfig(ETH_USDC_POOL, swapData, false)

	instructions, err := createCommonInstructions(ctx, params, config)
	if err != nil {
		return nil, err
	}

	closeAccountIx, err := createCloseAccountInstruction(params)
	if err != nil {
		return nil, err
	}
	instructions = append(instructions, closeAccountIx)

	return instructions, nil
}
func CreateFullSwapInstructionsFromUsdcToSol(ctx context.Context, params token.SwapInstructions) ([]solana.Instruction, error) {
	swapData := make([]byte, 24)
	copy(swapData[0:8], []byte{0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8})
	binary.LittleEndian.PutUint64(swapData[8:16], params.Amount)
	binary.LittleEndian.PutUint64(swapData[16:24], 0)

	config := createSwapConfig(SOL_USDC_POOL, swapData, false)

	return createCommonInstructions(ctx, params, config)
}

func CreateFullSwapInstructionsFromSolToUsdc(ctx context.Context, params token.SwapInstructions) ([]solana.Instruction, error) {
	swapData := make([]byte, 24)
	copy(swapData[0:8], []byte{0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8})
	binary.LittleEndian.PutUint64(swapData[8:16], params.Amount)
	binary.LittleEndian.PutUint64(swapData[16:24], 0)

	config := createSwapConfig(SOL_USDC_POOL, swapData, false)

	return createCommonInstructions(ctx, params, config)
}

func DetermineSwapDirection(firstToken, secondToken solana.PublicKey) (string, error) {
	switch {
	case firstToken.Equals(USDC) && secondToken.Equals(WETH):
		return "USDC_TO_ETH", nil
	case firstToken.Equals(WETH) && secondToken.Equals(USDC):
		return "ETH_TO_USDC", nil
	case firstToken.Equals(USDC) && secondToken.Equals(WSOL):
		return "USDC_TO_SOL", nil
	case firstToken.Equals(WSOL) && secondToken.Equals(USDC):
		return "SOL_TO_USDC", nil
	default:
		return "", fmt.Errorf("unsupported token pair: %s -> %s", firstToken, secondToken)
	}
}

func ConvertToRawAmount(amount float64, decimals int) uint64 {
	multiplier := math.Pow(10, float64(decimals))
	rawAmount := amount * multiplier
	return uint64(rawAmount)
}

func GetTokenDecimals(token solana.PublicKey) int {
	info, exists := tokenInfo[token.String()]
	if !exists {
		return 6
	}
	return int(info.decimals)
}

func ExecuteSwap(ctx context.Context, client *rpc.Client, params SwapParams) (solana.Signature, error) {
	swapDirection, err := DetermineSwapDirection(params.FromToken, params.ToToken)
	if err != nil {
		return solana.Signature{}, err
	}

	var instructions []solana.Instruction
	swapParams := token.SwapInstructions{
		Payer:       params.Wallet,
		FirstToken:  params.FromToken,
		SecondToken: params.ToToken,
		Amount:      ConvertToRawAmount(params.Amount, GetTokenDecimals(params.FromToken)),
		IsETH:       params.IsETH,
	}

	switch swapDirection {
	case "USDC_TO_ETH":
		instructions, err = CreateFullSwapInstructionsFromUsdcToEth(ctx, swapParams)
	case "ETH_TO_USDC":
		instructions, err = CreateFullSwapInstructionsFromEthToUsdc(ctx, swapParams)
	case "USDC_TO_SOL":
		instructions, err = CreateFullSwapInstructionsFromUsdcToSol(ctx, swapParams)
	case "SOL_TO_USDC":
		instructions, err = CreateFullSwapInstructionsFromSolToUsdc(ctx, swapParams)
	default:
		return solana.Signature{}, fmt.Errorf("unsupported swap direction: %s", swapDirection)
	}

	if err != nil {
		return solana.Signature{}, fmt.Errorf("error creating swap instructions: %v", err)
	}

	return ExecuteTransaction(ctx, client, instructions, params.Wallet)
}

func ExecuteTransaction(ctx context.Context, client *rpc.Client, instructions []solana.Instruction, feePayer solana.PrivateKey) (solana.Signature, error) {
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

	log.Printf("Transaction sent succesfully: %s%s", constants.EclipseScan, sig)

	return sig, nil
}
