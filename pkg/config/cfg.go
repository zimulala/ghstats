package config

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

// Config contains configuration options.
type Config struct {
	PTAL `toml:"ptal"`
}

// Access contains access token for services.
type Access struct {
	// Github access token
	GithubToken string `toml:"github-token"`
	// Feishu webhook bot
	FeishuWebhookToken string `toml:"feishu-webhook-token"`
}

// PTAL contains configuration options for PTAL command.
type PTAL struct {
	Access  `toml:"access"`
	PRQuery []string `toml:"pr-query"`
}

// ReadConfig reads config for config file.
func ReadConfig(cfgPath string) (*Config, error) {
	b, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := toml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
