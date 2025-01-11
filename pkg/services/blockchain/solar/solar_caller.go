package solar

import (
	"context"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/base"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/blockchain/lifinity"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/balance"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	httpClient http.Client,
	client *rpc.Client,
	cfg configs.AppConfig,
	acc *model.EclipseAccount,
	notifier *telegram.Notifier,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Solar Swap")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if cfg.Modules.Mode == "eth" {
			amountDecimals, err := balance.GetUSDCBalance(ctx, client, acc.PublicKey)
			if err != nil {
				return false, fmt.Errorf("error getting USDC balance: %v", err)
			}

			if amountDecimals == 0 {
				return false, fmt.Errorf("USDC balance is 0")
			}

			logger.Info("Пытаюсь выполнить свап всего баланса USDC -> ETH")

			params := token.SwapInstructions{
				Payer:         acc.PrivateKey,
				FirstToken:    lifinity.USDC,
				SecondToken:   lifinity.WETH,
				Amount:        amountDecimals,
				IsETH:         false,
				TokenSymbol:   "USDC",
				TokenDecimals: 6,
			}

			err = balance.CheckAndWaitForBalance(ctx, client, params, amountDecimals, maxAttempts, cfg.MinEthHold)
			if err != nil {
				logger.Error("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			swapParams := SwapParams{
				Amount:    strconv.FormatUint(amountDecimals, 10),
				FromToken: lifinity.USDC,
				ToToken:   lifinity.WETH,
			}

			swapResponse, err := GetSolarSwapCompute(httpClient, swapParams)
			if err != nil {
				logger.Error("Ошибка получения compute (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			destinationATA, _, err := token.FindAssociatedTokenAddress2022(acc.PublicKey, lifinity.USDC)
			if err != nil {
				return false, fmt.Errorf("error getting ATA: %v", err)
			}

			txResponse, err := CreateSwapTransaction(httpClient, acc.PublicKey.String(), destinationATA.String(), swapResponse)
			if err != nil {
				logger.Error("Ошибка создания транзакции (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			sig, err := ExecuteSwapFromInstructions(ctx, client, txResponse.Data[0].Transaction, acc.PrivateKey)
			if err != nil {
				logger.Error("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				time.Sleep(3 * time.Second)
				continue
			}

			notifier.AddSuccessMessageWithTxLink(
				acc.PublicKey.String(),
				"Solar Swap: All USDC -> ETH",
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		} else {
			firstPair, secondPair, tokenType, err := base.GetRandomTokenPair(cfg.Invariant.Tokens)
			if err != nil {
				return false, err
			}

			var value float64
			var valueStr string

			switch tokenType {
			case "ETH":
				value, valueStr = randomizer.GetRandomValueWithPrecision(
					cfg.Invariant.Native.ETH.MinValue,
					cfg.Invariant.Native.ETH.MaxValue,
					cfg.Invariant.Native.ETH.MinPrecision,
					cfg.Invariant.Native.ETH.MaxPrecision,
					float64(firstPair.Decimals),
				)
			case "SOL":
				value, valueStr = randomizer.GetRandomValueWithPrecision(
					cfg.Invariant.Native.SOL.MinValue,
					cfg.Invariant.Native.SOL.MaxValue,
					cfg.Invariant.Native.SOL.MinPrecision,
					cfg.Invariant.Native.SOL.MaxPrecision,
					float64(firstPair.Decimals),
				)
			case "stable":
				value, valueStr = randomizer.GetRandomValueWithPrecision(
					cfg.Invariant.Stable.MinValue,
					cfg.Invariant.Stable.MaxValue,
					cfg.Invariant.Stable.MinPrecision,
					cfg.Invariant.Stable.MaxPrecision,
					float64(firstPair.Decimals),
				)
			default:
				return false, fmt.Errorf("unknown token type: %s", tokenType)
			}

			amountDecimals, err := strconv.ParseUint(valueStr, 10, 64)
			if err != nil {
				return false, fmt.Errorf("error parsing value string: %v", err)
			}

			logger.Info("Пытаюсь выполнить свап %f %s -> %s", value, firstPair.Symbol, secondPair.Symbol)

			params := token.SwapInstructions{
				Payer:         acc.PrivateKey,
				FirstToken:    firstPair.Address,
				SecondToken:   secondPair.Address,
				Amount:        amountDecimals,
				IsETH:         firstPair.Address.String() == lifinity.WETH.String(),
				TokenSymbol:   firstPair.Symbol,
				TokenDecimals: firstPair.Decimals,
			}

			err = balance.CheckAndWaitForBalance(ctx, client, params, amountDecimals, maxAttempts, cfg.MinEthHold)
			if err != nil {
				logger.Error("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			swapParams := SwapParams{
				Amount:    valueStr,
				FromToken: firstPair.Address,
				ToToken:   secondPair.Address,
			}

			swapResponse, err := GetSolarSwapCompute(httpClient, swapParams)
			if err != nil {
				return false, fmt.Errorf("error getting swap compute: %v", err)
			}

			destinationATA, _, err := token.FindAssociatedTokenAddress2022(acc.PublicKey, lifinity.USDC)

			if err != nil {
				return false, fmt.Errorf("error getting ATA: %v", err)
			}

			txResponse, err := CreateSwapTransaction(httpClient, acc.PublicKey.String(), destinationATA.String(), swapResponse)
			if err != nil {
				return false, fmt.Errorf("error creating swap transaction: %v", err)
			}

			if sig, err := ExecuteSwapFromInstructions(ctx, client, txResponse.Data[0].Transaction, acc.PrivateKey); err != nil {
				logger.Error("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				time.Sleep(3 * time.Second)
				continue
			} else {
				notifier.AddSuccessMessageWithTxLink(
					acc.PublicKey.String(),
					fmt.Sprintf("Solar Swap: %.6f %s -> %s", value, firstPair.Symbol, secondPair.Symbol),
					constants.EclipseScan,
					sig.String(),
				)
				return true, nil
			}
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Solar Swap")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
