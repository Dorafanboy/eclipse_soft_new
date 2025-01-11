package orca

import (
	"context"
	"database/sql"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/base"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/interfaces"
	"eclipse/pkg/services/blockchain/lifinity"
	"eclipse/pkg/services/database"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/balance"
	"fmt"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
)

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	rpcClient *rpc.Client,
	cfg configs.AppConfig,
	acc *model.EclipseAccount,
	proxyManager interfaces.ProxyManagerInterface,
	notifier *telegram.Notifier,
	db *sql.DB,
	accountIndex int,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Orca Swap")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if cfg.Modules.Mode == "eth" {
			amountDecimals, err := balance.GetUSDCBalance(ctx, rpcClient, acc.PublicKey)
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

			err = balance.CheckAndWaitForBalance(ctx, rpcClient, params, amountDecimals, maxAttempts, cfg.MinEthHold)
			if err != nil {
				logger.Error("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			quoteParams := SwapQuoteParams{
				FromToken:            lifinity.USDC.String(),
				ToToken:              lifinity.WETH.String(),
				Amount:               strconv.FormatUint(amountDecimals, 10),
				IsLegacy:             true,
				AmountIsInput:        true,
				IncludeData:          true,
				IncludeComputeBudget: false,
				MaxTxSize:            1200,
				WalletAddress:        acc.PublicKey.String(),
			}

			proxy := proxyManager.GetProxyURL(accountIndex)

			resp, err := GetOrcaSwapQuote(quoteParams, proxy)
			if err != nil {
				logger.Error("Ошибка получения котировки (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			instructions, err := PrepareSwapInstructions(resp, acc.PublicKey.String(), proxy)
			if err != nil {
				logger.Error("Ошибка подготовки инструкций (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			sig, err := SimulateAndSendTransaction(ctx, rpcClient, instructions, acc.PrivateKey)
			if err != nil {
				logger.Error("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
				time.Sleep(3 * time.Second)
				continue
			}

			amount := float64(amountDecimals) / 1_000_000

			if db != nil && cfg.Database.Enabled {
				err = database.AddModule(
					db,
					acc.PublicKey.String(),
					"Orca",
					fmt.Sprintf("%.6f", amount),
					"USDC",
					sig.String(),
				)
				if err != nil {
					logger.Error("Failed to add module to database: %v", err)
				}
			}

			notifier.AddSuccessMessageWithTxLink(
				acc.PublicKey.String(),
				"Orca Swap: All USDC -> ETH",
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		} else {
			var value float64
			var valueStr string

			firstPair, secondPair, tokenType, err := base.GetRandomTokenPair(cfg.Invariant.Tokens)
			if err != nil {
				return false, err
			}

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

			isETH := firstPair.Address.String() == lifinity.WETH.String()

			params := token.SwapInstructions{
				Payer:         acc.PrivateKey,
				FirstToken:    firstPair.Address,
				SecondToken:   secondPair.Address,
				Amount:        amountDecimals,
				IsETH:         isETH,
				TokenSymbol:   firstPair.Symbol,
				TokenDecimals: firstPair.Decimals,
			}

			quoteParams := SwapQuoteParams{
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

			logger.Info("Пытаюсь выполнить свап %f %s -> %s", value, firstPair.Symbol, secondPair.Symbol)

			err = balance.CheckAndWaitForBalance(ctx, rpcClient, params, amountDecimals, maxAttempts, cfg.MinEthHold)
			if err != nil {
				logger.Error("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			proxy := proxyManager.GetProxyURL(accountIndex)

			resp, err := GetOrcaSwapQuote(quoteParams, proxy)
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
				formatString := fmt.Sprintf("%%.%df", firstPair.Decimals)

				if db != nil && cfg.Database.Enabled {
					err = database.AddModule(
						db,
						acc.PublicKey.String(),
						"Orca",
						fmt.Sprintf(formatString, value),
						firstPair.Symbol,
						sig.String(),
					)
					if err != nil {
						logger.Error("Failed to add module to database: %v", err)
					}
				}

				notifier.AddSuccessMessageWithTxLink(
					acc.PublicKey.String(),
					fmt.Sprintf("Orca Swap: %.6f %s -> %s", value, firstPair.Symbol, secondPair.Symbol),
					constants.EclipseScan,
					sig.String(),
				)
				return true, nil
			}
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Orca Swap")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
