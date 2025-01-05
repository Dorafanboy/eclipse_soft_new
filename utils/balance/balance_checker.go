package balance

import (
	"context"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"math"
	"math/big"
	"strconv"
	"time"
)

func GetTokenBalance(ctx context.Context, client *rpc.Client, params token.SwapInstructions) (uint64, error) {
	if params.IsETH {
		balance, err := client.GetBalance(
			ctx,
			params.Payer.PublicKey(),
			rpc.CommitmentFinalized,
		)
		if err != nil {
			return 0, fmt.Errorf("ошибка получения ETH баланса: %v", err)
		}

		return balance.Value, nil
	} else {
		tokenAccount, _, err := token.FindAssociatedTokenAddress2022(
			params.Payer.PublicKey(),
			params.FirstToken,
		)
		if err != nil {
			return 0, fmt.Errorf("ошибка получения токена аккаунта: %v", err)
		}

		balance, err := client.GetTokenAccountBalance(
			ctx,
			tokenAccount,
			rpc.CommitmentFinalized,
		)
		if err != nil {
			return 0, fmt.Errorf("ошибка получения баланса аккаунта: %v", err)
		}

		amount, err := strconv.ParseUint(balance.Value.Amount, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("ошибка парсинга баланса аккаунта: %v", err)
		}

		return amount, nil
	}
}

func CheckAndWaitForBalance(ctx context.Context, client *rpc.Client, params token.SwapInstructions, requiredAmount uint64, maxAttempts int) error {
	tokenName := "ETH"
	decimals := float64(9)

	if !params.IsETH {
		tokenName = params.TokenSymbol
		decimals = float64(params.TokenDecimals)
	}

	reqFloat := new(big.Float).SetUint64(requiredAmount)
	divisor := new(big.Float).SetFloat64(math.Pow(10, decimals))
	humanReq := new(big.Float).Quo(reqFloat, divisor)

	var reqFloatVal float64
	reqFloatVal, _ = humanReq.Float64()

	for i := 0; i < maxAttempts; i++ {
		balance, err := GetTokenBalance(ctx, client, params)
		if err != nil {
			logger.Error("Ошибка проверки баланса %s (попытка %d/%d): %v", tokenName, i+1, maxAttempts, err)
			time.Sleep(time.Second * 3)
			continue
		}

		balanceFloat := new(big.Float).SetUint64(balance)
		humanBalance := new(big.Float).Quo(balanceFloat, divisor)

		var floatVal float64
		floatVal, _ = humanBalance.Float64()

		if balance >= requiredAmount {
			logger.Info("Баланс %s найден: %.6f (требуется: %.6f)",
				tokenName, floatVal, reqFloatVal)
			return nil
		}

		time.Sleep(time.Second * 1)
	}

	return fmt.Errorf("не найден баланс %s после %d попыток", tokenName, maxAttempts)
}
