package main

import (
	"eclipse/cmd"
	"eclipse/configs"
	"eclipse/internal/logger"
	"eclipse/pkg/services/file"
	"eclipse/pkg/services/telegram"
	"eclipse/storage"
	"eclipse/utils/format"
	"eclipse/utils/managers"
	"eclipse/utils/shuffle"
	"fmt"
	"time"
)

func main() {
	if err := run(); err != nil {
		logger.Error(err.Error())
	}
}

func run() error {
	appCfg, err := configs.NewAppConfig()
	if err != nil {
		return err
	}

	if appCfg.IsShuffle {
		logger.Info("Включен режим перемешивания кошельков")

		err = shuffle.ShuffleFiles("../data/evm_private_keys.txt", "../data/eclipse_private_keys.txt")
		if err != nil {
			return err
		}
	}

	wallets, err := storage.LoadWallets("../data/evm_private_keys.txt", "../data/eclipse_private_keys.txt")
	if err != nil {
		return err
	}

	wordLists, err := file.LoadWordsFromFile("../words/words.txt")
	if err != nil {
		logger.Error("Error loading words: %v", err)
		return err
	}

	proxies, err := file.ReadLines("../data/proxies.txt")
	if err != nil {
		return err
	}

	if len(proxies) == 0 {
		return fmt.Errorf("надо указать как минимум одно прокси для работы скрипта")
	}

	evmWallets := wallets.EvmAccounts
	eclipseWallets := wallets.Eclipse

	logger.Info("Успешно подгрузил %d, EVM кошельков, %d, ECLIPSE кошельков, %d прокси", len(evmWallets), len(eclipseWallets), len(proxies))

	proxyManager := managers.NewProxyManager(proxies, len(wallets.EvmAccounts))

	logger.Info("Успешно подгрузил конфиг")
	if appCfg.Threads.Enabled {
		logger.Info("Включен режим многопоточного запуска аккаунтов")
	}

	if appCfg.Telegram.Enabled {
		logger.Info("Включен режим отправки уведомлений в телеграм")
	}

	notifier, err := telegram.NewNotifier(appCfg.Telegram.BotToken, appCfg.Telegram.UserID)
	if err != nil {
		return err
	}

	logger.Info("Включенные модули")
	logger.Info("━━━━━━━━━━━━━━━━━━━━━━")

	logger.Info("• Orca:      %v", format.FormatStatus(appCfg.Modules.Enabled.Orca))
	logger.Info("• Lifinity:  %v", format.FormatStatus(appCfg.Modules.Enabled.Lifinity))
	logger.Info("• Invariant: %v", format.FormatStatus(appCfg.Modules.Enabled.Invariant))
	logger.Info("• Relay:     %v", format.FormatStatus(appCfg.Modules.Enabled.Relay))
	logger.Info("• Solar:     %v", format.FormatStatus(appCfg.Modules.Enabled.Solar))
	logger.Info("• Underdog:  %v", format.FormatStatus(appCfg.Modules.Enabled.Underdog))

	logger.Info("━━━━━━━━━━━━━━━━━━━━━━\n")

	logger.Info("Ожидаю 10 секунд для просмотра включенных модулей и начинаю")

	time.Sleep(time.Second * 10)

	moduleManager := managers.NewModuleManager(*appCfg.Modules)

	return cmd.StartSoft(*wallets, *appCfg, moduleManager, proxyManager, notifier, wordLists)
}
