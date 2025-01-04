package configs

import (
	"eclipse/constants"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type TelegramConfig struct {
	Enabled  bool   `yaml:"enabled"`
	BotToken string `yaml:"bot_token"`
	UserID   int64  `yaml:"user_id"`
}

func NewTelegramConfig() (*TelegramConfig, error) {
	data, err := os.ReadFile(constants.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading telegram_config: %v", err)
	}

	var wrapper struct {
		Telegram TelegramConfig `yaml:"telegram"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("error unmarshaling telegram_config: %v", err)
	}

	return &wrapper.Telegram, nil
}
