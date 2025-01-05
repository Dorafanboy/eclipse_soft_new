package orca

import (
	"context"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/base"
	"eclipse/internal/logger"
	"eclipse/model"
	"eclipse/pkg/interfaces"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"time"
)

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	rpcClient *rpc.Client,
	cfg configs.InvariantConfig,
	acc *model.EclipseAccount,
	proxyManager interfaces.ProxyManagerInterface,
	notifier *telegram.Notifier,
	accountIndex int,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Orca Swap")

	var value float64
	var valueStr string

	for attempt := 0; attempt < maxAttempts; maxAttempts++ {
		firstPair, secondPair, tokenType, err := base.GetRandomTokenPair(cfg.Tokens)
		if err != nil {
			return false, err
		}

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

		params := SwapQuoteParams{
			FromToken:            firstPair.Address.String(),
			ToToken:              secondPair.Address.String(),
			Amount:               valueStr,
			IsLegacy:             true,
			AmountIsInput:        true,
			IncludeData:          true,
			IncludeComputeBudget: false,
			MaxTxSize:            1200,
			WalletAddress:        acc.PublicKey.String(),
		}

		proxy := proxyManager.GetProxyURL(accountIndex)

		resp, err := GetOrcaSwapQuote(params, proxy)
		if err != nil {
			return false, err
		}

		instructions, err := PrepareSwapInstructions(resp, acc.PublicKey.String(), proxy)
		if err != nil {
			return false, err
		}

		sig, err := SimulateAndSendTransaction(ctx, rpcClient, instructions, acc.PrivateKey)
		if err != nil {
			logger.Error("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		} else {
			notifier.AddSuccessMessageWithTxLink(
				acc.PublicKey.String(),
				fmt.Sprintf("Orca Swap: %.6f %s -> %s", value, firstPair.Symbol, secondPair.Symbol),
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Orca Swap")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
