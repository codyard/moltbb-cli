package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"moltbb-cli/internal/utils"
)

const DefaultAPIBaseURL = "https://api.moltbb.com"

type Config struct {
	APIBaseURL            string   `yaml:"api_base_url"`
	AllowInsecureHTTP     bool     `yaml:"allow_insecure_http,omitempty"`
	InputPaths            []string `yaml:"input_paths"`
	OutputDir             string   `yaml:"output_dir"`
	Template              string   `yaml:"template,omitempty"`
	SyncOnRun             bool     `yaml:"sync_on_run"`
	RequestTimeoutSeconds int      `yaml:"request_timeout_seconds"`
	RetryCount            int      `yaml:"retry_count"`
	OpenClawLogPath       string   `yaml:"openclaw_log_path,omitempty"`
	DiariesDir            string   `yaml:"diaries_dir,omitempty"`
}

func Default() Config {
	return Config{
		APIBaseURL:            DefaultAPIBaseURL,
		InputPaths:            []string{"~/.openclaw/logs/work.log"},
		OutputDir:             "diary",
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
	if strings.HasPrefix(c.APIBaseURL, "https://") {
		c.AllowInsecureHTTP = false
	} else if strings.HasPrefix(c.APIBaseURL, "http://") {
		if !c.AllowInsecureHTTP {
			return fmt.Errorf("api_base_url must use https unless allow_insecure_http is true: %s", c.APIBaseURL)
		}
	} else {
		return fmt.Errorf("api_base_url must start with https:// or http://: %s", c.APIBaseURL)
	}

	if len(c.InputPaths) == 0 {
		if strings.TrimSpace(c.OpenClawLogPath) != "" {
			c.InputPaths = []string{c.OpenClawLogPath}
		} else {
			c.InputPaths = append([]string{}, Default().InputPaths...)
		}
	}

	normalizedInputPaths := make([]string, 0, len(c.InputPaths))
	for _, p := range c.InputPaths {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		expanded, err := utils.ExpandPath(trimmed)
		if err != nil {
			return fmt.Errorf("normalize input_paths entry: %w", err)
		}
		normalizedInputPaths = append(normalizedInputPaths, filepath.Clean(expanded))
	}
	if len(normalizedInputPaths) == 0 {
		return fmt.Errorf("input_paths cannot be empty")
	}
	c.InputPaths = normalizedInputPaths

	if strings.TrimSpace(c.OutputDir) == "" {
		if strings.TrimSpace(c.DiariesDir) != "" {
			c.OutputDir = c.DiariesDir
		} else {
			c.OutputDir = Default().OutputDir
		}
	}

	outputDir, err := utils.ExpandPath(c.OutputDir)
	if err != nil {
		return fmt.Errorf("normalize output_dir: %w", err)
	}
	c.OutputDir = filepath.Clean(outputDir)
	c.Template = strings.TrimSpace(c.Template)
	c.OpenClawLogPath = c.InputPaths[0]
	c.DiariesDir = c.OutputDir

	if c.RequestTimeoutSeconds <= 0 {
		c.RequestTimeoutSeconds = Default().RequestTimeoutSeconds
	}
	if c.RetryCount < 0 {
		c.RetryCount = Default().RetryCount
	}
	return nil
}

func ParseInputPathsCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
