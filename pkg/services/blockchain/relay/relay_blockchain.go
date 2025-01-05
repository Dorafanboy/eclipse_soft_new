package relay

import (
	"context"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/logger"
	"eclipse/model"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
	"strconv"
)

func MakeRelayBridge(ctx context.Context, acc model.EvmAccount, chainData configs.Chain, txData TransactionData) (common.Hash, error) {
	logger.Info("Произвожу вызов функции для Relay бриджа")

	client, err := ethclient.Dial(chainData.RPC)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to connect to the %s client: %v", chainData.Name, err)
	}

	nonce, err := client.PendingNonceAt(ctx, acc.Address)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to get nonce: %v", err)
	}

	head, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to get header: %v", err)
	}

	baseFee := head.BaseFee
	if baseFee == nil {
		return constants.ZeroHash, fmt.Errorf("base fee is nil, network might not support EIP-1559")
	}

	suggestedGasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to get suggested gas price: %v", err)
	}

	maxPriorityFeePerGas := new(big.Int).Sub(suggestedGasPrice, baseFee)
	if maxPriorityFeePerGas.Cmp(big.NewInt(0)) <= 0 {
		maxPriorityFeePerGas = big.NewInt(1000000000)
	}

	maxFeePerGas := new(big.Int).Mul(big.NewInt(2), baseFee)
	maxFeePerGas.Add(maxFeePerGas, maxPriorityFeePerGas)

	to := common.HexToAddress(txData.To)

	feeCap, err := strconv.Atoi(txData.MaxFeePerGas)
	tipCap, err := strconv.Atoi(txData.MaxPriorityFeePerGas)
	value, err := strconv.Atoi(txData.Value)

	data := common.FromHex(txData.Data)

	msg := ethereum.CallMsg{
		From:      acc.Address,
		To:        &to,
		GasFeeCap: big.NewInt(int64(feeCap)),
		GasTipCap: big.NewInt(int64(tipCap)),
		Value:     big.NewInt(int64(value)),
		Data:      data,
	}

	gasLimit, err := client.EstimateGas(ctx, msg)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("transaction simulation failed: %v", err)
	}

	balance, err := client.BalanceAt(ctx, acc.Address, nil)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to get balance: %v", err)
	}

	maxGasCost := new(big.Int).Mul(maxFeePerGas, big.NewInt(int64(gasLimit)))
	totalCost := new(big.Int).Add(maxGasCost, big.NewInt(int64(value)))

	if balance.Cmp(totalCost) < 0 {
		return constants.ZeroHash, fmt.Errorf("insufficient balance: have %v need %v", balance, totalCost)
	}

	var gasPrice *big.Int
	if txData.MaxFeePerGas == "0" || txData.MaxFeePerGas == "" {
		priorityFee, err := strconv.ParseInt(txData.MaxPriorityFeePerGas, 10, 64)
		if err != nil {
			return constants.ZeroHash, fmt.Errorf("failed to parse MaxPriorityFeePerGas: %v", err)
		}
		gasPrice = big.NewInt(priorityFee)
	} else {
		maxFee, err := strconv.ParseInt(txData.MaxFeePerGas, 10, 64)
		if err != nil {
			return constants.ZeroHash, fmt.Errorf("failed to parse MaxFeePerGas: %v", err)
		}
		gasPrice = big.NewInt(maxFee)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &to,
		Value:    big.NewInt(int64(value)),
		Data:     data,
	})

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(int64(chainData.ChainID))), acc.PrivateKey)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to sign tx: %v", err)
	}

	logger.Success("✅ Симуляция успешна! Отправляем транзакцию...")

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return constants.ZeroHash, fmt.Errorf("failed to send transaction: %v", err)
	}

	hash := signedTx.Hash()

	logger.Success("Транзакция успешно отправлена %s%s\n", chainData.ScanURL, hash)
	fmt.Println()

	return hash, nil
}
