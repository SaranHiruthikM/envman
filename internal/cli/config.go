package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	configDirName  = ".envman"
	configFileName = "config.json"
	projFileName   = ".envman.json"
)

// GlobalConfig lives at ~/.envman/config.json (0600).
// PrivateKey is the user's decrypted X25519 private key, base64-encoded.
// It is sensitive and must never leave this file.
type GlobalConfig struct {
	Token      string `json:"token"`
	ApiURL     string `json:"api_url"`
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// ProjectConfig lives at ./.envman.json and is committed to git.
// Contains no secrets.
type ProjectConfig struct {
	ProjectID   string `json:"project_id"`
	Environment string `json:"env"`
	ServerURL   string `json:"server_url,omitempty"`
}

func globalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName, configFileName), nil
}

func globalConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName), nil
}

func SaveGlobalConfig(cfg *GlobalConfig) error {
	dir, err := globalConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, configFileName)
	return os.WriteFile(path, jsonData, 0600)
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	path, err := globalConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func DeleteGlobalConfig() error {
	path, err := globalConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func projectConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, projFileName), nil
}

func SaveProjectConfig(cfg *ProjectConfig) error {
	path, err := projectConfigPath()
	if err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

func LoadProjectConfig() (*ProjectConfig, error) {
	path, err := projectConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.ProjectID == "" {
		return nil, errors.New("project config is missing project_id")
	}
	return &cfg, nil
}
