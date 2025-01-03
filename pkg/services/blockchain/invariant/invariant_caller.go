package invariant

import (
	"context"
	"eclipse/configs"
	"eclipse/internal/base"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/randomizer"
	"eclipse/utils/balance"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	WRAPPED_ETH_ADDRESS = "So11111111111111111111111111111111111111112"
)

type Module struct{}

func (m *Module) Execute(ctx context.Context, httpClient http.Client, rpcClient *rpc.Client, cfg configs.InvariantConfig, acc *model.EclipseAccount, maxAttempts int) (bool, error) {
	log.Println("Начал выполнение модуля Invariant Swap")

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

		isETH := firstPair.Address.String() == WRAPPED_ETH_ADDRESS

		log.Printf("Пытаюсь выполнить свап %f %s -> %s", value, firstPair.Symbol, secondPair.Symbol)

		params := token.SwapInstructions{
			Payer:         acc.PrivateKey,
			FirstToken:    firstPair.Address,
			SecondToken:   secondPair.Address,
			Amount:        amountDecimals,
			IsETH:         isETH,
			TokenSymbol:   firstPair.Symbol,
			TokenDecimals: firstPair.Decimals,
		}

		err = balance.CheckAndWaitForBalance(ctx, rpcClient, params, amountDecimals, 3)
		if err != nil {
			log.Printf("Insufficient balance for pair (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			continue
		}

		instructions, newAccountKeypair, err := CreateFullSwapInstructions(params)
		if err != nil {
			return false, fmt.Errorf("error creating instructions: %v", err)
		}

		_, err = InvariantSendTx(ctx, rpcClient, instructions, params.Payer, newAccountKeypair)
		if err != nil {
			log.Printf("Ошибка свапа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			continue
		} else {
			return true, nil
		}
	}

	return false, fmt.Errorf("could not find sufficient balance after %d attempts", maxAttempts)
}
