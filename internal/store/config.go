package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"port_sentinel/internal/ports"
)

type UIConfig struct {
	AutoRefreshEnabled    bool `json:"autoRefreshEnabled"`
	AutoRefreshIntervalMs int  `json:"autoRefreshIntervalMs"`
	ForceKillEnabled      bool `json:"forceKillEnabled"`
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
			AutoRefreshEnabled:    false,
			AutoRefreshIntervalMs: 5000,
			ForceKillEnabled:      false,
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil && !errors.Is(err, os.ErrPermission) {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	return nil
}
