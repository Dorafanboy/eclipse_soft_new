package gas_station

import (
	"context"
	"database/sql"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/database"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/balance"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

var usdc = "AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE"

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	httpClient http.Client,
	rpcClient *rpc.Client,
	cfg configs.AppConfig,
	acc *model.EclipseAccount,
	notifier *telegram.Notifier,
	db *sql.DB,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Gas Station")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		value, valueStr := randomizer.GetRandomValueWithPrecision(
			cfg.Invariant.Stable.MinValue,
			cfg.Invariant.Stable.MaxValue,
			cfg.Invariant.Stable.MinPrecision,
			cfg.Invariant.Stable.MaxPrecision,
			float64(6),
		)

		amountDecimals, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return false, fmt.Errorf("error parsing value string: %v", err)
		}

		logger.Info("Пытаюсь выполнить Gas Station %f USDC -> ETH", value)

		params := token.SwapInstructions{
			Payer:         acc.PrivateKey,
			FirstToken:    solana.MustPublicKeyFromBase58(usdc),
			Amount:        amountDecimals,
			TokenSymbol:   "USDC",
			TokenDecimals: 6,
		}

		err = balance.CheckAndWaitForBalance(ctx, rpcClient, params, amountDecimals, maxAttempts, cfg.MinEthHold)
		if err != nil {
			logger.Error("Insufficient balance for USDC (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			continue
		}

		amount, err := strconv.Atoi(valueStr)
		if err != nil {
			return false, fmt.Errorf("error converting value to int: %v", err)
		}

		request := SwapRequest{
			User:              acc.PublicKey.String(),
			SourceMint:        usdc,
			Amount:            amount,
			SlippingTolerance: 1,
		}

		response, err := GetTxData(httpClient, request)
		if err != nil {
			log.Fatal(err)
		}

		sig, err := SendTransaction(ctx, rpcClient, acc.PrivateKey, response.Transaction)
		if err != nil {
			logger.Error("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		} else {
			if db != nil && cfg.Database.Enabled {
				err = database.AddModule(
					db,
					acc.PublicKey.String(),
					"Gas Station",
					fmt.Sprintf("%.6f", value),
					"USDC",
					sig.String(),
				)
				if err != nil {
					logger.Error("Failed to add module to database: %v", err)
				}
			}
			notifier.AddSuccessMessageWithTxLink(
				acc.PublicKey.String(),
				fmt.Sprintf("Gas Station: %.6f USDC -> ETH", value),
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Gas Station")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
