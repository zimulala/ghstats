package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pelletier/go-toml"
)

const (
	githubTokenEnvKey        = "GHSTATS_GITHUB_TOKEN"
	feishuWebhookTokenEnvKey = "GHSTATS_FEISHU_WEBHOOK_TOKEN"
)

// Config contains configuration options.
type Config struct {
	PTAL           `toml:"ptal"` // ptal and pkgs all use this configure.
	Review         `toml:"review"`
	IsOnlyPrintMsg bool `toml:"print-msg-local"` // Check whether the message is only printed locally.
}

// Access contains access token for services.
type Access struct {
	// Github access token
	GithubToken string `toml:"github-token"`
	// Feishu webhook bot
	FeishuWebhookToken string `toml:"feishu-webhook-token"`
}

// If the Access struct has no value, get it from the environment variables.
// This is useful for runing a cron with GitHub Actions without exposing the cfg.
func (a *Access) getFromEnv() {
	if len(a.GithubToken) == 0 {
		a.GithubToken = os.Getenv(githubTokenEnvKey)
	}
	if len(a.FeishuWebhookToken) == 0 {
		a.FeishuWebhookToken = os.Getenv(feishuWebhookTokenEnvKey)
	}
}

// Repo contains configuration options for Repo in PTAL command.
type Repo struct {
	Name        string   `toml:"name"`
	PRQuery     []string `toml:"pr-query"`
	Packages    []string `toml:"allow-pkgs"`
	PROwnerRepo string   `toml:"pr-owner-repo"`
}

// PTAL contains configuration options for PTAL command.
type PTAL struct {
	Access     `toml:"access"`
	ReportName string `toml:"report-name"`
	Repos      []Repo `toml:"repos"`
}

func (ptal PTAL) ReposName() string {
	names := make([]string, 0, len(ptal.Repos))
	for _, repo := range ptal.Repos {
		names = append(names, repo.Name)
	}
	return strings.Join(names, ", ")
}

type Review struct {
	Access        `toml:"access"`
	Repos         []Repo   `toml:"repos"`
	LGTMComments  []string `toml:"lgtm-comments"`
	BlockComments []string `toml:"block-comments"`
	AllowUsers    []string `toml:"allow-users"`
	BlockUsers    []string `toml:"block-users"`
	BlockLabels   []string `toml:"block-labels"`
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
	cfg.PTAL.Access.getFromEnv()
	cfg.Review.Access.getFromEnv()
	return cfg, nil
}
