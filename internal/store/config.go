package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"port_sentinel/internal/ports"
)

type UIConfig struct {
	AutoRefreshEnabled   bool `json:"autoRefreshEnabled"`
	AutoRefreshIntervalMs int  `json:"autoRefreshIntervalMs"`
	ForceKillEnabled     bool `json:"forceKillEnabled"`
}

type Config struct {
	PresetPorts map[int]bool `json:"presetPorts"`
	CustomPorts []int        `json:"customPorts"`
	PinnedPorts map[int]bool `json:"pinnedPorts"`
	UI          UIConfig     `json:"ui"`
}

func DefaultConfig() Config {
	return Config{
		PresetPorts: ports.DefaultPresetPorts(),
		CustomPorts: []int{},
		PinnedPorts: map[int]bool{},
		UI: UIConfig{
			AutoRefreshEnabled:   false,
			AutoRefreshIntervalMs: 5000,
			ForceKillEnabled:     false,
		},
	}
}

func ConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "portsentinel", "config.json"), nil
}

func LoadConfig() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig(), err
	}
	if cfg.PresetPorts == nil {
		cfg.PresetPorts = ports.DefaultPresetPorts()
	}
	if cfg.CustomPorts == nil {
		cfg.CustomPorts = []int{}
	}
	if cfg.PinnedPorts == nil {
		cfg.PinnedPorts = map[int]bool{}
	}
	if cfg.UI.AutoRefreshIntervalMs == 0 {
		cfg.UI.AutoRefreshIntervalMs = 5000
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
