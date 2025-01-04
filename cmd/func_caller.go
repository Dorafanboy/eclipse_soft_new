package cmd

import (
	"context"
	"eclipse/configs"
	"eclipse/internal/base"
	"eclipse/pkg/interfaces"
	"eclipse/pkg/services/blockchain/relay"
	"eclipse/pkg/services/file"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/storage"
	"eclipse/utils/managers"
	"fmt"
	"github.com/gagliardetto/solana-go/rpc"
	"log"
	"math/rand"
	"sync"
)

func StartSoft(wallets storage.WalletStorage, cfg configs.AppConfig, moduleManager *managers.ModuleManager, proxyManager *managers.ProxyManager, notifier *telegram.Notifier, lists *file.WordLists) error {
	ctx := context.Background()

	if !cfg.Threads.Enabled {
		return processAccountsRange(ctx, 0, 0, len(wallets.EvmAccounts), wallets, cfg, moduleManager, proxyManager, notifier, lists)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, cfg.Threads.Count)

	accountsPerThread := len(wallets.EvmAccounts) / cfg.Threads.Count
	if accountsPerThread == 0 {
		accountsPerThread = 1
	}

	for i := 0; i < cfg.Threads.Count; i++ {
		start := i * accountsPerThread
		end := start + accountsPerThread
		if i == cfg.Threads.Count-1 {
			end = len(wallets.EvmAccounts)
		}

		wg.Add(1)
		go func(threadNum, start, end int) {
			defer wg.Done()
			if err := processAccountsRange(ctx, threadNum, start, end, wallets, cfg, moduleManager, proxyManager, notifier, lists); err != nil {
				errChan <- err
			}
		}(i, start, end)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func processAccountsRange(ctx context.Context, threadNum, start, end int, wallets storage.WalletStorage, cfg configs.AppConfig, moduleManager *managers.ModuleManager, proxyManager *managers.ProxyManager, notifier *telegram.Notifier, lists *file.WordLists) error {
	var moduleNames []string
	for name := range moduleManager.EnabledModules {
		moduleNames = append(moduleNames, name)
	}

	var res bool
	var err error

	for i := start; i < end; i++ {
		httpClient := proxyManager.GetHttpClient(i)
		eclipseAcc := wallets.Eclipse[i]
		evmAcc := wallets.EvmAccounts[i]

		rpcClient := rpc.New("https://mainnetbeta-rpc.eclipse.xyz")

		notifier.AddMessageForWallet(eclipseAcc.PublicKey.String(),
			fmt.Sprintf("[%d/%d]\nEVM: %s \nECLIPSE: %s",
				i+1,
				len(wallets.EvmAccounts),
				evmAcc.Address.String(),
				eclipseAcc.PublicKey.String(),
			),
		)

		log.Printf("[Thread %d] Account [%d/%d] start EVM: %s, ECLIPSE: %s\n\n",
			threadNum+1,
			i+1,
			len(wallets.EvmAccounts),
			evmAcc.Address.String(),
			eclipseAcc.PublicKey.String(),
		)

		moduleInfo := moduleManager.EnabledModules["Relay"]

		if cfg.Modules.Enabled.Relay {
			relayModule := relay.Module{}
			res, err = relayModule.Execute(
				ctx,
				*cfg.Relay,
				evmAcc,
				eclipseAcc,
				rpcClient,
				*httpClient,
				notifier,
				cfg.Delay.BetweenRetries.Attempts,
			)

			if err != nil {
				log.Printf("[Thread %d] Error: %v", threadNum+1, err)
			}

			if err == nil || !res {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false) // тут потом поменять true/false
			} else {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false) // тут потом поменять true/false
			}
		}

		numModules := base.GetRandomPrecision(
			cfg.Modules.ModulesCount.Min,
			cfg.Modules.ModulesCount.Max,
		)

		fmt.Println()
		log.Printf("Буду выполнять %d модулей на аккаунте %s\n", numModules, eclipseAcc.PublicKey.String())

		for j := 0; j < numModules; {
			randomIndex := rand.Intn(len(moduleNames))
			moduleName := moduleNames[randomIndex]

			if moduleName == "Relay" {
				continue
			}

			fmt.Println()
			moduleInfo = moduleManager.EnabledModules[moduleName]

			switch moduleInfo.Type {
			case interfaces.OrcaType:
				module := moduleInfo.Module.(interfaces.OrcaModule)
				res, err = module.Execute(ctx, rpcClient, *cfg.Invariant, eclipseAcc, proxyManager, notifier, i, cfg.Delay.BetweenRetries.Attempts)

			case interfaces.UnderdogType:
				module := moduleInfo.Module.(interfaces.UnderdogModule)
				res, err = module.Execute(ctx, *httpClient, rpcClient, eclipseAcc, notifier, lists.Words, cfg.Delay.BetweenRetries.Attempts)

			case interfaces.DefaultType:
				module := moduleInfo.Module.(interfaces.DefaultModule)
				res, err = module.Execute(ctx, *httpClient, rpcClient, *cfg.Invariant, eclipseAcc, notifier, cfg.Delay.BetweenRetries.Attempts)
			}
			j++

			if err == nil || !res {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false)
			} else if j < numModules-1 {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false)
			}
		}

		err = notifier.SendWalletMessages(eclipseAcc.PublicKey.String())
		if err != nil {
			log.Printf("Ошибка отправки сообщений: %v", err)
		} else {
			log.Printf("Сообщения успешно отправлены")
		}

		if i+1 == len(wallets.Eclipse) {
			log.Println("Все аккаунты отработаны")
			return nil
		}

		if res {
			log.Printf("[Thread %d] Accounts [%d/%d] EVM: %s, ECLIPSE: %s successfully ended\n\n",
				threadNum+1,
				i+1, len(wallets.EvmAccounts),
				wallets.EvmAccounts[i].Address.String(),
				wallets.Eclipse[i].PublicKey.String(),
			)
			randomizer.RandomDelay(cfg.Delay.BetweenAccounts.Min, cfg.Delay.BetweenAccounts.Max, true)
		} else {
			log.Printf("[Thread %d] Accounts [%d/%d] EVM: %s, ECLIPSE: %s ended with errors\n\n",
				threadNum+1,
				i+1, len(wallets.EvmAccounts),
				wallets.EvmAccounts[i].Address.String(),
				wallets.Eclipse[i].PublicKey.String(),
			)
			randomizer.RandomDelay(cfg.Delay.BetweenAccounts.Min, cfg.Delay.BetweenAccounts.Max, false)
		}
	}

	return nil
}
