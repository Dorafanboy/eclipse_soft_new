package interfaces

import (
	"context"
	"eclipse/configs"
	"eclipse/model"
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
		maxAttempts int,
	) (bool, error)
}
