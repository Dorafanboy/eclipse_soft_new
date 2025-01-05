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
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"net/http"
	"time"
)

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	httpClient http.Client,
	client *rpc.Client,
	cfg configs.InvariantConfig,
	acc *model.EclipseAccount,
	notifier *telegram.Notifier,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Solar Swap")
	for attempt := 0; attempt < maxAttempts; attempt++ {
		firstPair, secondPair, tokenType, err := base.GetRandomTokenPair(cfg.Tokens)
		if err != nil {
			return false, err
		}

		var value float64
		var valueStr string

		switch tokenType {
		case "ETH":
			value, valueStr = randomizer.GetRandomValueWithPrecision(
				cfg.Native.ETH.MinValue,
				cfg.Native.ETH.MaxValue,
				cfg.Native.ETH.MinPrecision,
				cfg.Native.ETH.MaxPrecision,
				float64(firstPair.Decimals),
			)
		case "SOL":
			value, valueStr = randomizer.GetRandomValueWithPrecision(
				cfg.Native.SOL.MinValue,
				cfg.Native.SOL.MaxValue,
				cfg.Native.SOL.MinPrecision,
				cfg.Native.SOL.MaxPrecision,
				float64(firstPair.Decimals),
			)
		case "stable":
			value, valueStr = randomizer.GetRandomValueWithPrecision(
				cfg.Stable.MinValue,
				cfg.Stable.MaxValue,
				cfg.Stable.MinPrecision,
				cfg.Stable.MaxPrecision,
				float64(firstPair.Decimals),
			)
		default:
			return false, fmt.Errorf("unknown token type: %s", tokenType)
		}

		logger.Info("Буду выполнять свап %f %s -> %s", value, firstPair.Symbol, secondPair.Symbol)

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

	notifier.AddErrorMessage(acc.PublicKey.String(), "Solar Swap")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
