package relay

import (
	"context"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/randomizer"
	"eclipse/utils/balance"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"math"
	"net/http"
	"time"
)

type Module struct{}

func (m *Module) Execute(ctx context.Context, cfg configs.RelayConfig, evmAccount *model.EvmAccount, eclipseAccount *model.EclipseAccount, rpcClient *rpc.Client, httpClient http.Client, maxAttempts int) (bool, error) {
	log.Println("Начал выполнение модуля Relay Bridge")

	for attempt := 0; attempt < maxAttempts; attempt++ {
		params := token.SwapInstructions{
			Payer:         eclipseAccount.PrivateKey,
			FirstToken:    solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112"),
			TokenSymbol:   "ETH",
			TokenDecimals: 9,
			IsETH:         true,
		}

		balance, err := balance.GetTokenBalance(ctx, rpcClient, params)
		if err != nil {
			return false, fmt.Errorf("ошибка при проверке баланса ETH: %v", err)
		}

		minBalanceFloat := cfg.EthBridge.MinBalance
		minBalanceWei := uint64(minBalanceFloat * math.Pow10(9))

		if balance >= minBalanceWei {
			log.Printf("Баланс ETH в Eclipse (%.6f) больше минимального (%.6f). Пропускаем бридж",
				float64(balance)/math.Pow10(9),
				minBalanceFloat)
			return false, nil
		}

		log.Printf("Баланс ETH в Eclipse (%.6f) меньше минимального (%.6f). Выполняю бридж",
			float64(balance)/math.Pow10(9),
			minBalanceFloat)

		valueWei, valueStr := randomizer.GetRandomValueWithPrecision(cfg.EthBridge.MinValue, cfg.EthBridge.MaxValue, cfg.EthBridge.MinPrecision, cfg.EthBridge.MaxPrecision, 18)

		randChain := configs.GetRandomChainFromNames(cfg.Networks.Chains)

		log.Printf("Буду выполнять бридж %s -> Eclipse, %f ETH", randChain.Name, valueWei)

		request := RelayRequest{
			User:                 evmAccount.Address.String(),
			OriginChainId:        randChain.ChainID,
			DestinationChainId:   constants.DestChainId,
			OriginCurrency:       constants.ZeroAddress.String(),
			DestinationCurrency:  constants.DestCurrency,
			Recipient:            eclipseAccount.PublicKey.String(),
			TradeType:            constants.TradeInput,
			Amount:               valueStr,
			Referrer:             constants.Referrer,
			UseExternalLiquidity: false,
		}

		response, err := GetRelayData(httpClient, request)
		if err != nil {
			return false, err
		}

		_, err = MakeRelayBridge(ctx, *evmAccount, randChain, *response)
		if err != nil {
			log.Printf("Ошибка бриджа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			fmt.Println()
			continue
		} else {
			return true, nil
		}
	}

	return true, nil
}
