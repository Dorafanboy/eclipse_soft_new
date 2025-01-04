package interfaces

import (
	"context"
	"eclipse/configs"
	"eclipse/model"
	"eclipse/pkg/services/telegram"
	"github.com/gagliardetto/solana-go/rpc"
	"net/http"
)

type ModuleType int

const (
	OrcaType ModuleType = iota
	UnderdogType
	DefaultType
)

type ModuleInfo struct {
	Module interface{}
	Type   ModuleType
}

type ProxyManagerInterface interface {
	GetProxyURL(accountIndex int) string
}

type OrcaModule interface {
	Execute(
		ctx context.Context,
		rpcClient *rpc.Client,
		cfg configs.InvariantConfig,
		acc *model.EclipseAccount,
		proxyManager ProxyManagerInterface,
		notifier *telegram.Notifier,
		accountIndex int,
		maxAttempts int,
	) (bool, error)
}

type UnderdogModule interface {
	Execute(
		ctx context.Context,
		httpClient http.Client,
		rpcClient *rpc.Client,
		acc *model.EclipseAccount,
		notifier *telegram.Notifier,
		words []string,
		maxAttempts int,
	) (bool, error)
}

type DefaultModule interface {
	Execute(
		ctx context.Context,
		httpClient http.Client,
		rpcClient *rpc.Client,
		cfg configs.InvariantConfig,
		acc *model.EclipseAccount,
		notifier *telegram.Notifier,
		maxAttempts int,
	) (bool, error)
}
