package store

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	cfg := DefaultConfig()
	cfg.CustomPorts = []int{1234, 4321}
	cfg.PinnedPorts = map[int]bool{80: true, 4321: true}
	cfg.UI.AutoRefreshEnabled = true
	cfg.UI.AutoRefreshIntervalMs = 2000
	cfg.UI.ForceKillEnabled = true

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Fatalf("expected temp dir to exist: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if !reflect.DeepEqual(cfg.CustomPorts, loaded.CustomPorts) {
		t.Fatalf("custom ports mismatch: %+v vs %+v", cfg.CustomPorts, loaded.CustomPorts)
	}
	if !reflect.DeepEqual(cfg.PinnedPorts, loaded.PinnedPorts) {
		t.Fatalf("pinned ports mismatch: %+v vs %+v", cfg.PinnedPorts, loaded.PinnedPorts)
	}
	if cfg.UI.AutoRefreshEnabled != loaded.UI.AutoRefreshEnabled {
		t.Fatalf("auto refresh mismatch")
	}
	if cfg.UI.AutoRefreshIntervalMs != loaded.UI.AutoRefreshIntervalMs {
		t.Fatalf("interval mismatch")
	}
	if cfg.UI.ForceKillEnabled != loaded.UI.ForceKillEnabled {
		t.Fatalf("force kill mismatch")
	}
}

func TestSaveConfigDoesNotLeaveTempFileAndReplacesContent(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APPDATA", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)

	first := DefaultConfig()
	first.CustomPorts = []int{1111}
	if err := SaveConfig(first); err != nil {
		t.Fatalf("first SaveConfig failed: %v", err)
	}

	second := DefaultConfig()
	second.CustomPorts = []int{2222, 3333}
	second.UI.AutoRefreshEnabled = true
	if err := SaveConfig(second); err != nil {
		t.Fatalf("second SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if !reflect.DeepEqual(second.CustomPorts, loaded.CustomPorts) {
		t.Fatalf("expected replaced custom ports: %+v vs %+v", second.CustomPorts, loaded.CustomPorts)
	}
	if loaded.UI.AutoRefreshEnabled != second.UI.AutoRefreshEnabled {
		t.Fatalf("expected replaced ui config")
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath failed: %v", err)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "config-") && strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("found unexpected temp config file: %s", entry.Name())
		}
	}
}
