package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDirPerm  os.FileMode = 0o700
	defaultFilePerm os.FileMode = 0o600
)

func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", errors.New("unable to resolve user home directory")
	}
	return home, nil
}

func ExpandPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is empty")
	}

	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := HomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}

	return filepath.Clean(path), nil
}

func MoltbbDir() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".moltbb"), nil
}

func ConfigPath() (string, error) {
	dir, err := MoltbbDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func CredentialsPath() (string, error) {
	dir, err := MoltbbDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func BindingPath() (string, error) {
	dir, err := MoltbbDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "binding.json"), nil
}

func EnsureDir(path string, perm os.FileMode) error {
	if perm == 0 {
		perm = defaultDirPerm
	}
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("create directory %s: %w", path, err)
	}
	return nil
}

func EnsureMoltbbDir() (string, error) {
	dir, err := MoltbbDir()
	if err != nil {
		return "", err
	}
	if err := EnsureDir(dir, defaultDirPerm); err != nil {
		return "", err
	}
	return dir, nil
}

func SecureWriteFile(path string, data []byte, perm os.FileMode) error {
	if perm == 0 {
		perm = defaultFilePerm
	}
	if err := EnsureDir(filepath.Dir(path), defaultDirPerm); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("set file permission %s: %w", path, err)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
