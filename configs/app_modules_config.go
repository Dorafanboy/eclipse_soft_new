package configs

import (
	"eclipse/constants"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ModulesCountConfig struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type EnabledModulesConfig struct {
	Orca       bool `yaml:"orca"`
	Lifinity   bool `yaml:"lifinity"`
	Invariant  bool `yaml:"invariant"`
	Relay      bool `yaml:"relay"`
	Solar      bool `yaml:"solar"`
	Underdog   bool `yaml:"underdog"`
	GasStation bool `yaml:"gas_station"`
}

type ModulesConfig struct {
	ModulesCount ModulesCountConfig   `yaml:"modules_count"`
	Mode         string               `yaml:"mode"`
	Sequence     []string             `yaml:"sequence"`
	Enabled      EnabledModulesConfig `yaml:"enabled"`
	Limited      struct {
		Underdog int `yaml:"underdog"`
	} `yaml:"limited"`
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
