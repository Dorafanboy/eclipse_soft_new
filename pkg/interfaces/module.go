﻿package interfaces

import (
	"context"
	"database/sql"
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
		cfg configs.AppConfig,
		acc *model.EclipseAccount,
		proxyManager ProxyManagerInterface,
		notifier *telegram.Notifier,
		db *sql.DB,
		accountIndex int,
		maxAttempts int,
	) (bool, error)
}

type UnderdogModule interface {
	Execute(
		ctx context.Context,
		httpClient http.Client,
		rpcClient *rpc.Client,
		cfg configs.AppConfig,
		acc *model.EclipseAccount,
		notifier *telegram.Notifier,
		db *sql.DB,
		words []string,
		minEthHold float64,
		maxAttempts int,
	) (bool, error)
}

type DefaultModule interface {
	Execute(
		ctx context.Context,
		httpClient http.Client,
		rpcClient *rpc.Client,
		cfg configs.AppConfig,
		acc *model.EclipseAccount,
		notifier *telegram.Notifier,
		db *sql.DB,
		maxAttempts int,
	) (bool, error)
}
