package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"moltbb-cli/internal/utils"
)

type Credentials struct {
	APIKey    string    `json:"api_key"`
	Token     string    `json:"token,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func Load() (Credentials, error) {
	path, err := utils.CredentialsPath()
	if err != nil {
		return Credentials{}, err
	}
	if !utils.FileExists(path) {
		return Credentials{}, errors.New("credentials file not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Credentials{}, fmt.Errorf("read credentials: %w", err)
	}

	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return Credentials{}, fmt.Errorf("parse credentials json: %w", err)
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return Credentials{}, errors.New("credentials missing api_key")
	}
	return c, nil
}

func Save(apiKey, token string) error {
	if strings.TrimSpace(apiKey) == "" {
		return errors.New("api key is empty")
	}

	path, err := utils.CredentialsPath()
	if err != nil {
		return err
	}

	c := Credentials{
		APIKey:    strings.TrimSpace(apiKey),
		Token:     strings.TrimSpace(token),
		UpdatedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	return utils.SecureWriteFile(path, data, 0o600)
}

func ResolveAPIKey() (string, error) {
	if env := strings.TrimSpace(os.Getenv("MOLTBB_API_KEY")); env != "" {
		return env, nil
	}
	c, err := Load()
	if err != nil {
		return "", err
	}
	return c.APIKey, nil
}
