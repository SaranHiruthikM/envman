package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	globalPath     = "/.envman"
	globalFileName = "/config.json"
	projFileName   = "/.envman.json"
)

type GlobalConfig struct {
	Token  string `json:"token"`
	ApiURL string `json:"api_url"`
}

type ProjectConfig struct {
	ProjectID   string `json:"project_id"`
	Environment string `json:"env"`
}

func SaveGlobalConfig(token, apiURL string) error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(userHomeDir+globalPath, 0755); err != nil {
		return err
	}

	data := map[string]string{
		"token":   token,
		"api_url": apiURL,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(userHomeDir, globalPath+globalFileName)

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return err
	}

	return nil
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(userHomeDir + globalPath + globalFileName)
	if err != nil {
		return nil, err
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil

}

func SaveProjectConfig(projectID, environment string) error {
	currDir, err := os.Getwd()
	if err != nil {
		return err
	}

	data := map[string]string{
		"project_id": projectID,
		"env":        environment,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(currDir, projFileName)

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return err
	}

	return nil

}
func LoadProjectConfig() (*ProjectConfig, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(currDir + projFileName)
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
