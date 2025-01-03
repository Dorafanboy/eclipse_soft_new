package configs

import (
	"eclipse/constants"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type ModulesCountConfig struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type EnabledModulesConfig struct {
	Orca      bool `yaml:"orca"`
	Lifinity  bool `yaml:"lifinity"`
	Invariant bool `yaml:"invariant"`
	Relay     bool `yaml:"relay"`
	Solar     bool `yaml:"solar"`
	Underdog  bool `yaml:"underdog"`
}

type ModulesConfig struct {
	ModulesCount ModulesCountConfig   `yaml:"modules_count"`
	Enabled      EnabledModulesConfig `yaml:"enabled"`
}

func NewModulesConfig() (*ModulesConfig, error) {
	data, err := os.ReadFile(constants.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading modules_config: %v", err)
	}

	var wrapper struct {
		Modules ModulesConfig `yaml:"modules"`
	}

	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("error unmarshaling modules_config: %v", err)
	}

	return &wrapper.Modules, nil
}
