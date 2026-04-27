package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	OutputDir string `json:"output_dir"`
}

var DefaultConfig = Config{
	OutputDir: ".", // Current directory by default
}

// LoadConfig reads the configuration from ~/.gotube/config.json
func LoadConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfig, nil // fallback to default if home dir not found
	}

	configPath := filepath.Join(homeDir, ".gotube", "config.json")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig, nil
		}
		return DefaultConfig, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return DefaultConfig, err
	}

	if cfg.OutputDir == "" {
		cfg.OutputDir = DefaultConfig.OutputDir
	}

	return cfg, nil
}

// SaveConfig saves the configuration to ~/.gotube/config.json
func SaveConfig(cfg Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".gotube")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cfg)
}
