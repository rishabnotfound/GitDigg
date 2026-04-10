package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Branch         string `mapstructure:"branch"`
	Output         string `mapstructure:"output"`
	Concurrency    int    `mapstructure:"concurrency"`
	Flat           bool   `mapstructure:"flat"`
	GitHubToken    string `mapstructure:"github_token"`
	GitLabToken    string `mapstructure:"gitlab_token"`
	BitbucketToken string `mapstructure:"bitbucket_token"`
}

func DefaultConfig() *Config {
	return &Config{
		Output:      ".",
		Concurrency: 4,
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	viper.SetConfigName(".gitdigg")
	viper.SetConfigType("yaml")

	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(home)
	}
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("GITDIG")
	viper.AutomaticEnv()

	viper.BindEnv("github_token", "GITHUB_TOKEN", "GH_TOKEN")
	viper.BindEnv("gitlab_token", "GITLAB_TOKEN")
	viper.BindEnv("bitbucket_token", "BITBUCKET_TOKEN")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) GetToken(provider string) string {
	switch provider {
	case "github":
		if c.GitHubToken != "" {
			return c.GitHubToken
		}
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			return token
		}
		return os.Getenv("GH_TOKEN")
	case "gitlab":
		if c.GitLabToken != "" {
			return c.GitLabToken
		}
		return os.Getenv("GITLAB_TOKEN")
	case "bitbucket":
		if c.BitbucketToken != "" {
			return c.BitbucketToken
		}
		return os.Getenv("BITBUCKET_TOKEN")
	}
	return ""
}

func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".gitdigg.yaml"
	}
	return filepath.Join(home, ".gitdigg.yaml")
}
