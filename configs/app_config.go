package configs

import (
	"eclipse/constants"
	"github.com/gagliardetto/solana-go"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Orca struct {
		Tokens []string     `yaml:"tokens"`
		Native NativeConfig `yaml:"native"`
		Stable AmountConfig `yaml:"stable"`
	} `yaml:"swaps"`
}

type ThreadConfig struct {
	Count   int  `yaml:"count"`
	Enabled bool `yaml:"enabled"`
}

type AppConfig struct {
	Orca      *OrcaConfig
	Invariant *InvariantConfig
	Relay     *RelayConfig
	Delay     *DelayConfig
	Modules   *ModulesConfig
	Threads   ThreadConfig `yaml:"threads"`
	Telegram  *TelegramConfig
	IsShuffle bool `yaml:"is_shuffle"`
}

type Token struct {
	Symbol   string
	Address  solana.PublicKey
	Decimals int8
}

type NativeConfig struct {
	ETH AmountConfig `yaml:"eth"`
	SOL AmountConfig `yaml:"sol"`
}

type AmountConfig struct {
	MinValue     float64 `yaml:"min_value"`
	MaxValue     float64 `yaml:"max_value"`
	MinPrecision int     `yaml:"min_precision"`
	MaxPrecision int     `yaml:"max_precision"`
}

func NewAppConfig() (*AppConfig, error) {
	orcaConfig, err := NewOrcaConfig()
	if err != nil {
		return nil, err
	}

	invariantConfig, err := NewInvariantConfig()
	if err != nil {
		return nil, err
	}

	relayConfig, err := NewRelayConfig()
	if err != nil {
		return nil, err
	}

	delayConfig, err := NewDelayConfig()
	if err != nil {
		return nil, err
	}

	modulesConfig, err := NewModulesConfig()
	if err != nil {
		return nil, err
	}

	telegramConfig, err := NewTelegramConfig()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(constants.ConfigPath)
	if err != nil {
		return nil, err
	}

	var wrapper struct {
		Threads   ThreadConfig `yaml:"threads"`
		IsShuffle bool         `yaml:"is_shuffle"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	return &AppConfig{
		Orca:      orcaConfig,
		Invariant: invariantConfig,
		Relay:     relayConfig,
		Delay:     delayConfig,
		Modules:   modulesConfig,
		Threads:   wrapper.Threads,
		Telegram:  telegramConfig,
		IsShuffle: wrapper.IsShuffle,
	}, nil
}

func LoadConfig() (*Config, error) {
	f, err := os.ReadFile(constants.ConfigPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
