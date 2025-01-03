package configs

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type NetworkConfig struct {
	Chains []string `yaml:"chains"`
}

type RelayConfig struct {
	EthBridge EthBridgeConfig `yaml:"eth_bridge"`
	Networks  NetworkConfig   `yaml:"networks"`
}

type EthBridgeConfig struct {
	MinBalance   float64 `yaml:"min_balance"`
	MinValue     float64 `yaml:"min_value"`
	MaxValue     float64 `yaml:"max_value"`
	MinPrecision int     `yaml:"min_precision"`
	MaxPrecision int     `yaml:"max_precision"`
}

func NewRelayConfig() (*RelayConfig, error) {
	data, err := os.ReadFile("../data/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("error when reading relay_config: %v", err)
	}

	var config RelayConfig

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshal to relay_config: %v", err)
	}

	return &config, nil
}
