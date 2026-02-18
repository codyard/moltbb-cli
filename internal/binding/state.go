package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"moltbb-cli/internal/utils"
)

type State struct {
	Bound            bool      `json:"bound"`
	BotID            string    `json:"bot_id,omitempty"`
	ActivationStatus string    `json:"activation_status,omitempty"`
	Hostname         string    `json:"hostname,omitempty"`
	OS               string    `json:"os,omitempty"`
	Version          string    `json:"version,omitempty"`
	Fingerprint      string    `json:"fingerprint,omitempty"`
	LastSyncAt       string    `json:"last_sync_at,omitempty"`
	LastSyncStatus   string    `json:"last_sync_status,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func Load() (State, error) {
	path, err := utils.BindingPath()
	if err != nil {
		return State{}, err
	}
	if !utils.FileExists(path) {
		return State{}, errors.New("binding not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, fmt.Errorf("read binding file: %w", err)
	}

	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, fmt.Errorf("parse binding json: %w", err)
	}
	return s, nil
}

func Save(s State) error {
	s.UpdatedAt = time.Now().UTC()
	path, err := utils.BindingPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal binding json: %w", err)
	}
	return utils.SecureWriteFile(path, data, 0o600)
}
