package configs

import (
	"eclipse/constants"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type DelayRange struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

type RetryConfig struct {
	Min      float64 `yaml:"min"`
	Max      float64 `yaml:"max"`
	Attempts int     `yaml:"attempts"`
}

type DelayConfig struct {
	BetweenAccounts DelayRange  `yaml:"between_accounts"`
	BetweenModules  DelayRange  `yaml:"between_modules"`
	BetweenRetries  RetryConfig `yaml:"between_retries"`
}

func NewDelayConfig() (*DelayConfig, error) {
	data, err := os.ReadFile(constants.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading delay_config: %v", err)
	}

	var wrapper struct {
		Delay DelayConfig `yaml:"delay"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("error unmarshaling delay_config: %v", err)
	}

	return &wrapper.Delay, nil
}
