package config

import (
	"os"
	"path/filepath"
)

const (
	// #nosec G101 - environment variable names, not credentials
	EnvOpenAIAPIKey = "OPENAI_API_KEY"
	EnvModel        = "AB_MODEL"
	EnvOpenABaseURL = "AB_OPENAI_BASE_URL"
	EnvKernelDir    = "AB_KERNEL_DIR"
	EnvStateDir     = "AB_STATE_DIR"
)

const (
	DefaultModel         = "gpt-4.1-mini"
	DefaultOpenAIModel   = "gpt-4.1-mini"
	DefaultOpenAIBaseURL = "https://api.openai.com/v1"
	DefaultStateDir      = "~/.local/state/ab"
)

type Config struct {
	OpenAIAPIKey  string
	Model         string
	OpenAIBaseURL string
	KernelDir     string
	StateDir      string
}

func Load() *Config {
	c := &Config{
		OpenAIAPIKey:  os.Getenv(EnvOpenAIAPIKey),
		Model:         os.Getenv(EnvModel),
		OpenAIBaseURL: os.Getenv(EnvOpenABaseURL),
		KernelDir:     os.Getenv(EnvKernelDir),
		StateDir:      os.Getenv(EnvStateDir),
	}

	if c.Model == "" {
		c.Model = DefaultModel
	}
	if c.OpenAIBaseURL == "" {
		c.OpenAIBaseURL = DefaultOpenAIBaseURL
	}
	if c.StateDir == "" {
		c.StateDir = expandHome(DefaultStateDir)
	}

	return c
}

func expandHome(path string) string {
	home := os.Getenv("HOME")
	if path == "~" {
		return home
	}
	if path == "~/" {
		return home
	}
	if len(path) > 1 && path[:2] == "~/" {
		return filepath.Join(home, path[2:])
	}
	return path
}

func (c *Config) StateDirPath() string {
	return expandHome(c.StateDir)
}

func (c *Config) KernelDirPath() string {
	return expandHome(c.KernelDir)
}
