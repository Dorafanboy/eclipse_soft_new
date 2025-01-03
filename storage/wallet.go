package storage

import (
	"eclipse/model"
	"eclipse/pkg/services/file"
	"errors"
	"fmt"
)

type WalletStorage struct {
	EvmAccounts []*model.EvmAccount
	Eclipse     []*model.EclipseAccount
}

func LoadWallets(evmKeysPath, eclipseKeysPath string) (*WalletStorage, error) {
	evmAccs, err := file.ReadAccounts(evmKeysPath, file.EVM)
	if err != nil {
		return nil, fmt.Errorf("failed to load EVM accounts: %w", err)
	}
	evmAccounts := evmAccs.([]*model.EvmAccount)

	eclipseAccs, err := file.ReadAccounts(eclipseKeysPath, file.ECLIPSE)
	if err != nil {
		return nil, fmt.Errorf("failed to load ECLIPSE accounts: %w", err)
	}
	eclipseAccounts := eclipseAccs.([]*model.EclipseAccount)

	if len(evmAccounts) != len(eclipseAccounts) {
		return nil, errors.New("number of evm accounts does not match the number of eclipse accounts")
	}

	return newWalletStorage(evmAccounts, eclipseAccounts), nil
}

func newWalletStorage(evmAccs []*model.EvmAccount, eclipseAccs []*model.EclipseAccount) *WalletStorage {
	return &WalletStorage{
		EvmAccounts: evmAccs,
		Eclipse:     eclipseAccs,
	}
}
