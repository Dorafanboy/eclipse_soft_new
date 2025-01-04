package lifinity

import (
	"context"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/base"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/balance"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"net/http"
	"strconv"
	"time"
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
	cfg configs.InvariantConfig,
	acc *model.EclipseAccount,
	notifier *telegram.Notifier,
	maxAttempts int,
) (bool, error) {
	log.Println("Начал выполнение модуля Lifinity Swap")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		firstPair, secondPair, tokenType, err := base.GetRandomTokenPair(cfg.Tokens)
		if err != nil {
			return false, fmt.Errorf("error getting token pair: %v", err)
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

		err = balance.CheckAndWaitForBalance(ctx, client, params, amountDecimals, 3)
		if err != nil {
			log.Printf("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			continue
		}

		log.Printf("Пытаюсь выполнить свап %f %s -> %s", value, firstPair.Symbol, secondPair.Symbol)

		swapParams := SwapParams{
			Amount:    value,
			FromToken: firstPair.Address,
			ToToken:   secondPair.Address,
			Wallet:    acc.PrivateKey,
			IsETH:     isETH,
		}

		sig, err := ExecuteSwap(ctx, client, swapParams)
		if err != nil {
			log.Printf("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		} else {
			notifier.AddSuccessMessageWithTxLink(
				acc.PublicKey.String(),
				fmt.Sprintf("Lifinity Swap: %.6f %s -> %s", value, firstPair.Symbol, secondPair.Symbol),
				constants.EclipseScan,
				sig.String(),
			)
			return true, nil
		}
	}

	notifier.AddErrorMessage(acc.PublicKey.String(), "Lifinity Swap")
	return false, fmt.Errorf("could not execute swap after %d attempts", maxAttempts)
}
