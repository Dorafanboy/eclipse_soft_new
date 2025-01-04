package main

import (
	"eclipse/cmd"
	"eclipse/configs"
	"eclipse/pkg/services/file"
	"eclipse/pkg/services/telegram"
	"eclipse/storage"
	"eclipse/utils/format"
	"eclipse/utils/managers"
	"log"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	wallets, err := storage.LoadWallets("../data/evm_private_keys.txt", "../data/eclipse_private_keys.txt")
	if err != nil {
		return err
	}

	wordLists, err := file.LoadWordsFromFile("../words/words.txt")
	if err != nil {
		log.Printf("Error loading words: %v", err)
		return err
	}

	proxies, err := file.ReadLines("../data/proxies.txt")
	if err != nil {
		return err
	}

	evmWallets := wallets.EvmAccounts
	eclipseWallets := wallets.Eclipse

	log.Printf("Успешно подгрузил %d, EVM кошельков, %d, ECLIPSE кошельков, %d прокси", len(evmWallets), len(eclipseWallets), len(proxies))

	proxyManager := managers.NewProxyManager(proxies, len(wallets.EvmAccounts))

	appCfg, err := configs.NewAppConfig()
	if err != nil {
		return err
	}

	log.Println("Успешно подгрузил конфиг")
	if appCfg.Threads.Enabled {
		log.Println("Включен режим многопоточного запуска аккаунтов")
	}

	if appCfg.Telegram.Enabled {
		log.Println("Включен режим отправки уведомлений в телеграм")
	}

	notifier, err := telegram.NewNotifier(appCfg.Telegram.BotToken, appCfg.Telegram.UserID)
	if err != nil {
		return err
	}

	log.Println("Включенные модули")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━")

	log.Printf("• Orca:      %v", format.FormatStatus(appCfg.Modules.Enabled.Orca))
	log.Printf("• Lifinity:  %v", format.FormatStatus(appCfg.Modules.Enabled.Lifinity))
	log.Printf("• Invariant: %v", format.FormatStatus(appCfg.Modules.Enabled.Invariant))
	log.Printf("• Relay:     %v", format.FormatStatus(appCfg.Modules.Enabled.Relay))
	log.Printf("• Solar:     %v", format.FormatStatus(appCfg.Modules.Enabled.Solar))
	log.Printf("• Underdog:  %v", format.FormatStatus(appCfg.Modules.Enabled.Underdog))

	log.Println("━━━━━━━━━━━━━━━━━━━━━━\n")

	log.Printf("Ожидаю 10 секунд для просмотра включенных модулей и начинаю")

	time.Sleep(time.Second * 10)

	moduleManager := managers.NewModuleManager(*appCfg.Modules)

	return cmd.StartSoft(*wallets, *appCfg, moduleManager, proxyManager, notifier, wordLists)
}
