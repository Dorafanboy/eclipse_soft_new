package cmd

import (
	"context"
	"database/sql"
	"eclipse/configs"
	"eclipse/internal/base"
	"eclipse/internal/logger"
	"eclipse/pkg/interfaces"
	"eclipse/pkg/services/blockchain/gas_station"
	"eclipse/pkg/services/blockchain/invariant"
	"eclipse/pkg/services/blockchain/lifinity"
	"eclipse/pkg/services/blockchain/orca"
	"eclipse/pkg/services/blockchain/relay"
	"eclipse/pkg/services/blockchain/solar"
	"eclipse/pkg/services/blockchain/underdog"
	"eclipse/pkg/services/file"
	"eclipse/pkg/services/randomizer"
	"eclipse/pkg/services/telegram"
	"eclipse/storage"
	"eclipse/utils/managers"
	"fmt"
	"math/rand"
	"sync"

	"github.com/gagliardetto/solana-go/rpc"
)

func StartSoft(wallets storage.WalletStorage, cfg configs.AppConfig, moduleManager *managers.ModuleManager, proxyManager *managers.ProxyManager, notifier *telegram.Notifier, db *sql.DB, lists *file.WordLists) error {
	ctx := context.Background()

	if !cfg.Threads.Enabled {
		return processAccountsRange(ctx, 0, 0, len(wallets.EvmAccounts), wallets, cfg, moduleManager, proxyManager, notifier, db, lists)
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
			if err := processAccountsRange(ctx, threadNum, start, end, wallets, cfg, moduleManager, proxyManager, notifier, db, lists); err != nil {
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

func processAccountsRange(ctx context.Context, threadNum, start, end int, wallets storage.WalletStorage, cfg configs.AppConfig, moduleManager *managers.ModuleManager, proxyManager *managers.ProxyManager, notifier *telegram.Notifier, db *sql.DB, lists *file.WordLists) error {
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

		logger.Info("[Thread %d] Account [%d/%d] start EVM: %s, ECLIPSE: %s\n\n",
			threadNum+1,
			i+1,
			len(wallets.EvmAccounts),
			evmAcc.Address.String(),
			eclipseAcc.PublicKey.String(),
		)

		if cfg.Modules.Mode == "random" && cfg.Modules.Enabled.Relay {
			relayModule := relay.Module{}
			res, err = relayModule.Execute(
				ctx,
				*cfg.Relay,
				evmAcc,
				eclipseAcc,
				rpcClient,
				*httpClient,
				notifier,
				db,
				cfg.Delay.BetweenRetries.Attempts,
			)

			if err != nil {
				logger.Error("[Thread %d] Error: %v", threadNum+1, err)
			}

			if err == nil || res {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, true)
			} else if err != nil {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false)
			}
		}

		fmt.Println()

		var modulesToExecute []string
		if cfg.Modules.Mode == "random" {
			availableModules := make([]string, 0)
			for name := range moduleManager.EnabledModules {
				if name != "Relay" {
					availableModules = append(availableModules, name)
				}
			}

			numModules := base.GetRandomPrecision(
				cfg.Modules.ModulesCount.Min,
				cfg.Modules.ModulesCount.Max,
			)

			for j := 0; j < numModules; j++ {
				randomIndex := rand.Intn(len(availableModules))
				modulesToExecute = append(modulesToExecute, availableModules[randomIndex])
			}

			logger.Info("Буду выполнять %d модулей на аккаунте %s\n", numModules, eclipseAcc.PublicKey.String())
		} else if cfg.Modules.Mode == "queue" {
			modulesToExecute = cfg.Modules.Sequence
			logger.Info("Буду выполнять %d модулей на аккаунте %s\n", len(modulesToExecute), eclipseAcc.PublicKey.String())
		} else if cfg.Modules.Mode == "eth" {
			swapModules := []string{"Orca", "Solar", "Invariant", "Lifinity"}
			randomModule := swapModules[rand.Intn(len(swapModules))]
			modulesToExecute = []string{randomModule}
			logger.Info("Буду выполнять %d модулей на аккаунте %s\n", len(modulesToExecute), eclipseAcc.PublicKey.String())
		}

		for moduleIndex, moduleName := range modulesToExecute {
			fmt.Println()
			logger.Info("Выполняю модуль %d/%d: %s", moduleIndex+1, len(modulesToExecute), moduleName)

			var moduleInfo interfaces.ModuleInfo
			if cfg.Modules.Mode == "random" {
				var exists bool
				moduleInfo, exists = moduleManager.EnabledModules[moduleName]
				if !exists {
					logger.Error("Модуль %s не найден", moduleName)
					continue
				}
			} else if cfg.Modules.Mode == "queue" {
				switch moduleName {
				case "Orca":
					moduleInfo = interfaces.ModuleInfo{
						Module: &orca.Module{},
						Type:   interfaces.OrcaType,
					}
				case "Underdog":
					moduleInfo = interfaces.ModuleInfo{
						Module: &underdog.Module{},
						Type:   interfaces.UnderdogType,
					}
				case "Invariant":
					moduleInfo = interfaces.ModuleInfo{
						Module: &invariant.Module{},
						Type:   interfaces.DefaultType,
					}
				case "Relay":
					moduleInfo = interfaces.ModuleInfo{
						Module: &relay.Module{},
						Type:   interfaces.DefaultType,
					}
				case "Lifinity":
					moduleInfo = interfaces.ModuleInfo{
						Module: &lifinity.Module{},
						Type:   interfaces.DefaultType,
					}
				case "Solar":
					moduleInfo = interfaces.ModuleInfo{
						Module: &solar.Module{},
						Type:   interfaces.DefaultType,
					}
				case "Gas Station":
					moduleInfo = interfaces.ModuleInfo{
						Module: &gas_station.Module{},
						Type:   interfaces.DefaultType,
					}
				default:
					logger.Error("Неизвестный модуль: %s", moduleName)
					continue
				}
			} else if cfg.Modules.Mode == "eth" {
				switch moduleName {
				case "Orca":
					moduleInfo = interfaces.ModuleInfo{
						Module: &orca.Module{},
						Type:   interfaces.OrcaType,
					}
				case "Solar":
					moduleInfo = interfaces.ModuleInfo{
						Module: &solar.Module{},
						Type:   interfaces.DefaultType,
					}
				case "Invariant":
					moduleInfo = interfaces.ModuleInfo{
						Module: &invariant.Module{},
						Type:   interfaces.DefaultType,
					}
				case "Lifinity":
					moduleInfo = interfaces.ModuleInfo{
						Module: &lifinity.Module{},
						Type:   interfaces.DefaultType,
					}
				}

				logger.Info("Режим ETH: используем %s для свапа USDC -> ETH", moduleName)

				if err != nil {
					logger.Error("[Thread %d] Error during ETH swap: %v", threadNum+1, err)
				}
			}

			switch moduleInfo.Type {
			case interfaces.OrcaType:
				module := moduleInfo.Module.(interfaces.OrcaModule)
				res, err = module.Execute(ctx, rpcClient, cfg, eclipseAcc, proxyManager, notifier, db, i, cfg.Delay.BetweenRetries.Attempts)

			case interfaces.UnderdogType:
				module := moduleInfo.Module.(interfaces.UnderdogModule)
				res, err = module.Execute(ctx, *httpClient, rpcClient, cfg, eclipseAcc, notifier, db, lists.Words, cfg.MinEthHold, cfg.Delay.BetweenRetries.Attempts)

			case interfaces.DefaultType:
				module := moduleInfo.Module.(interfaces.DefaultModule)
				res, err = module.Execute(ctx, *httpClient, rpcClient, cfg, eclipseAcc, notifier, db, cfg.Delay.BetweenRetries.Attempts)
			}

			if moduleIndex == len(modulesToExecute)-1 {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false)
			} else if err == nil || res {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, true)
			} else {
				randomizer.RandomDelay(cfg.Delay.BetweenModules.Min, cfg.Delay.BetweenModules.Max, false)
			}
		}

		err = notifier.SendWalletMessages(eclipseAcc.PublicKey.String())
		if err != nil {
			logger.Error("Ошибка отправки сообщений: %v", err)
		} else {
			logger.Info("Сообщения успешно отправлены в телеграм")
		}

		if i+1 == len(wallets.Eclipse) {
			fmt.Println()
			logger.Info("Все аккаунты отработаны")
			return nil
		}

		if res {
			logger.Info("[Thread %d] Accounts [%d/%d] EVM: %s, ECLIPSE: %s successfully ended\n\n",
				threadNum+1,
				i+1, len(wallets.EvmAccounts),
				wallets.EvmAccounts[i].Address.String(),
				wallets.Eclipse[i].PublicKey.String(),
			)
			randomizer.RandomDelay(cfg.Delay.BetweenAccounts.Min, cfg.Delay.BetweenAccounts.Max, true)
		} else {
			logger.Info("[Thread %d] Accounts [%d/%d] EVM: %s, ECLIPSE: %s ended with errors\n\n",
				threadNum+1,
				i+1, len(wallets.EvmAccounts),
				wallets.EvmAccounts[i].Address.String(),
				wallets.Eclipse[i].PublicKey.String(),
			)
			randomizer.RandomDelay(cfg.Delay.BetweenAccounts.Min, cfg.Delay.BetweenAccounts.Max, true)
		}
	}

	return nil
}
