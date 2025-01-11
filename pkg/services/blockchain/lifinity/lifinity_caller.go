package lifinity

import (
	"context"
	"database/sql"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/base"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/database"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/balance"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	USDC = solana.MustPublicKeyFromBase58("AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE")
	WETH = solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	WSOL = solana.MustPublicKeyFromBase58("BeRUj3h7BqkbdfFU7FBNYbodgf8GCHodzKvF9aVjNNfL")
)

var tokenInfo = map[string]struct {
	decimals int8
	symbol   string
}{
	"AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE": {6, "USDC"},
	"So11111111111111111111111111111111111111112":  {9, "ETH"},
	"BeRUj3h7BqkbdfFU7FBNYbodgf8GCHodzKvF9aVjNNfL": {9, "SOL"},
}

type SwapParams struct {
	Amount    float64
	FromToken solana.PublicKey
	ToToken   solana.PublicKey
	Wallet    solana.PrivateKey
	IsETH     bool
}

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	httpClient http.Client,
	client *rpc.Client,
	cfg configs.AppConfig,
	acc *model.EclipseAccount,
	notifier *telegram.Notifier,
	db *sql.DB,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Lifinity Swap")

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

			swapParams := SwapParams{
				FromToken: USDC,
				ToToken:   WETH,
				Wallet:    acc.PrivateKey,
				IsETH:     false,
			}

			sig, err := ExecuteSwap(ctx, client, swapParams)
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
					"Lifinity",
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
				"Lifinity Swap: All USDC -> ETH",
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		} else {
			firstPair, secondPair, tokenType, err := base.GetRandomTokenPair(cfg.Invariant.Tokens)
			if err != nil {
				return false, fmt.Errorf("error getting token pair: %v", err)
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

			isETH := firstPair.Address.String() == WETH.String()

			params := token.SwapInstructions{
				Payer:         acc.PrivateKey,
				FirstToken:    firstPair.Address,
				SecondToken:   secondPair.Address,
				Amount:        amountDecimals,
				IsETH:         isETH,
				TokenSymbol:   firstPair.Symbol,
				TokenDecimals: firstPair.Decimals,
			}

			logger.Info("Пытаюсь выполнить свап %f %s -> %s", value, firstPair.Symbol, secondPair.Symbol)

			err = balance.CheckAndWaitForBalance(ctx, client, params, amountDecimals, maxAttempts, cfg.MinEthHold)
			if err != nil {
				logger.Error("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
				continue
			}

			swapParams := SwapParams{
				Amount:    value,
				FromToken: firstPair.Address,
				ToToken:   secondPair.Address,
				Wallet:    acc.PrivateKey,
				IsETH:     isETH,
			}

			sig, err := ExecuteSwap(ctx, client, swapParams)
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
						"Lifinity",
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
					fmt.Sprintf("Lifinity Swap: %.6f %s -> %s", value, firstPair.Symbol, secondPair.Symbol),
					constants.EclipseScan,
					sig.String(),
				)
				return true, nil
			}
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Lifinity Swap")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
