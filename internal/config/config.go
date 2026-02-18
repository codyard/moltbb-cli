package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"moltbb-cli/internal/utils"
)

const DefaultAPIBaseURL = "https://api.moltbb.com"

type Config struct {
	APIBaseURL            string `yaml:"api_base_url"`
	OpenClawLogPath       string `yaml:"openclaw_log_path"`
	DiariesDir            string `yaml:"diaries_dir"`
	SyncOnRun             bool   `yaml:"sync_on_run"`
	RequestTimeoutSeconds int    `yaml:"request_timeout_seconds"`
	RetryCount            int    `yaml:"retry_count"`
}

func Default() Config {
	return Config{
		APIBaseURL:            DefaultAPIBaseURL,
		OpenClawLogPath:       "~/.openclaw/logs/work.log",
		DiariesDir:            "~/.moltbb/diaries",
		SyncOnRun:             true,
		RequestTimeoutSeconds: 12,
		RetryCount:            2,
	}
}

func Load() (Config, error) {
	path, err := utils.ConfigPath()
	if err != nil {
		return Config{}, err
	}
	if !utils.FileExists(path) {
		return Config{}, fmt.Errorf("config file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}

	if err := cfg.Normalize(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(cfg Config) error {
	if err := cfg.Normalize(); err != nil {
		return err
	}

	path, err := utils.ConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config yaml: %w", err)
	}

	return utils.SecureWriteFile(path, data, 0o600)
}

func Ensure() (Config, bool, error) {
	path, err := utils.ConfigPath()
	if err != nil {
		return Config{}, false, err
	}

	if utils.FileExists(path) {
		cfg, loadErr := Load()
		return cfg, false, loadErr
	}

	cfg := Default()
	if err := Save(cfg); err != nil {
		return Config{}, false, err
	}
	return cfg, true, nil
}

func (c *Config) Normalize() error {
	if strings.TrimSpace(c.APIBaseURL) == "" {
		c.APIBaseURL = DefaultAPIBaseURL
	}
	c.APIBaseURL = strings.TrimRight(strings.TrimSpace(c.APIBaseURL), "/")
	if !strings.HasPrefix(c.APIBaseURL, "https://") {
		return fmt.Errorf("api_base_url must use https: %s", c.APIBaseURL)
	}

	if strings.TrimSpace(c.OpenClawLogPath) == "" {
		c.OpenClawLogPath = Default().OpenClawLogPath
	}
	if strings.TrimSpace(c.DiariesDir) == "" {
		c.DiariesDir = Default().DiariesDir
	}

	openClaw, err := utils.ExpandPath(c.OpenClawLogPath)
	if err != nil {
		return fmt.Errorf("normalize openclaw_log_path: %w", err)
	}
	c.OpenClawLogPath = openClaw

	diaries, err := utils.ExpandPath(c.DiariesDir)
	if err != nil {
		return fmt.Errorf("normalize diaries_dir: %w", err)
	}
	c.DiariesDir = diaries

	if c.RequestTimeoutSeconds <= 0 {
		c.RequestTimeoutSeconds = Default().RequestTimeoutSeconds
	}
	if c.RetryCount < 0 {
		c.RetryCount = Default().RetryCount
	}
	return nil
}
