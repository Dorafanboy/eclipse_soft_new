package relay

import (
	"context"
	"eclipse/configs"
	"eclipse/constants"
	"eclipse/internal/logger"
	"eclipse/internal/token"
	"eclipse/model"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/utils/balance"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"math"
	"net/http"
	"time"
)

type Module struct{}

func (m *Module) Execute(
	ctx context.Context,
	cfg configs.RelayConfig,
	evmAccount *model.EvmAccount,
	eclipseAccount *model.EclipseAccount,
	rpcClient *rpc.Client,
	httpClient http.Client,
	notifier *telegram.Notifier,
	maxAttempts int,
) (bool, error) {
	logger.Info("Начал выполнение модуля Relay Bridge")

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
			logger.Info("Баланс ETH в Eclipse (%.6f) больше минимального (%.6f). Пропускаем бридж",
				float64(balance)/math.Pow10(9),
				minBalanceFloat)
			return false, nil
		}

		logger.Info("Баланс ETH в Eclipse (%.6f) меньше минимального (%.6f). Выполняю бридж",
			float64(balance)/math.Pow10(9),
			minBalanceFloat)

		valueWei, valueStr := randomizer.GetRandomValueWithPrecision(cfg.EthBridge.MinValue, cfg.EthBridge.MaxValue, cfg.EthBridge.MinPrecision, cfg.EthBridge.MaxPrecision, 18)

		randChain := configs.GetRandomChainFromNames(cfg.Networks.Chains)

		logger.Info("Буду выполнять бридж %s -> Eclipse, %f ETH", randChain.Name, valueWei)

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

		sig, err := MakeRelayBridge(ctx, *evmAccount, randChain, *response)
		if err != nil {
			logger.Error("Ошибка бриджа (попытка %d/%d): %v", attempt+1, maxAttempts, err)
			time.Sleep(3 * time.Second)
			fmt.Println()
			continue
		} else {
			notifier.AddSuccessMessageWithTxLink(
				eclipseAccount.PublicKey.String(),
				fmt.Sprintf("Relay Bridge: %s -> Eclipse, %f ETH", randChain.Name, valueWei),
				randChain.ScanURL,
				sig.String(),
			)
			return true, nil
		}
	}

	notifier.AddErrorMessage(eclipseAccount.PublicKey.String(), "Relay Bridge")
	return false, fmt.Errorf("could not execute bridge after %d attempts", maxAttempts)
}
