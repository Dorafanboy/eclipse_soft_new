package main

import (
	"database/sql"
	"eclipse/cmd"
	"eclipse/configs"
	"eclipse/internal/logger"
	"eclipse/pkg/services/database"
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

	var db *sql.DB

	if appCfg.Database.Enabled {
		logger.Info("Включен режим с использование базы данных")
		db, err = database.InitDB()
		if err != nil {
			logger.Error("Failed to init database: %v", err)
			return err
		}
		defer db.Close()
		logger.Success("База данных успешно инициализирована")
	}

	notifier, err := telegram.NewNotifier(appCfg.Telegram.BotToken, appCfg.Telegram.UserID)
	if err != nil {
		return err
	}

	if appCfg.Modules.Mode == "random" {
		logger.Info("Включен режим рандомного запуска модулей")
		logger.Info("Включенные модули:")
		logger.Info("━━━━━━━━━━━━━━━━━━━━━━")
		logger.Info("• Orca:         %v", format.FormatStatus(appCfg.Modules.Enabled.Orca))
		logger.Info("• Lifinity:     %v", format.FormatStatus(appCfg.Modules.Enabled.Lifinity))
		logger.Info("• Invariant:    %v", format.FormatStatus(appCfg.Modules.Enabled.Invariant))
		logger.Info("• Relay:        %v", format.FormatStatus(appCfg.Modules.Enabled.Relay))
		logger.Info("• Solar:        %v", format.FormatStatus(appCfg.Modules.Enabled.Solar))
		logger.Info("• Underdog:     %v", format.FormatStatus(appCfg.Modules.Enabled.Underdog))
		logger.Info("• Gas Station:  %v", format.FormatStatus(appCfg.Modules.Enabled.GasStation))
	} else if appCfg.Modules.Mode == "eth" {
		logger.Info("Включен режим ETH свапов(будут свапы всех балансов в ETH через рандомные свапалки ")
	} else {
		logger.Info("Включен режим последовательного запуска модулей")
		logger.Info("Последовательность выполнения модулей:")
		logger.Info("━━━━━━━━━━━━━━━━━━━━━━")

		for i, moduleName := range appCfg.Modules.Sequence {
			logger.Info("• %d. %s", i+1, moduleName)
		}
	}

	logger.Info("━━━━━━━━━━━━━━━━━━━━━━\n")

	logger.Info("Ожидаю 10 секунд для просмотра включенных модулей и начинаю")

	time.Sleep(time.Second * 10)

	moduleManager := managers.NewModuleManager(*appCfg.Modules)

	return cmd.StartSoft(*wallets, *appCfg, moduleManager, proxyManager, notifier, db, wordLists)
}
