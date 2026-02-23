package app

import (
	"errors"

	"port_sentinel/internal/ports"
	"port_sentinel/internal/store"
)

func RefreshAll(state *State) ([]ports.PortScanResult, error) {
	portsList := state.GetPorts()
	results, err := ports.ScanPorts(portsList)
	if len(results) > 0 {
		state.SetResults(results)
	}
	return results, err
}

func RefreshOne(state *State, port int) (ports.PortScanResult, error) {
	res, err := ports.ScanPort(port)
	if err == nil {
		state.SetResult(res)
	}
	return res, err
}

func KillProcess(pid int, force bool) error {
	return ports.KillPID(pid, force)
}

func SaveConfig(state *State) error {
	state.mu.Lock()
	cfg := state.Config
	state.mu.Unlock()
	return store.SaveConfig(cfg)
}

func UpdateUIConfig(state *State, updater func(cfg *store.Config) error) error {
	state.mu.Lock()
	cfg := state.Config
	state.mu.Unlock()
	if err := updater(&cfg); err != nil {
		return err
	}
	state.UpdateConfig(cfg)
	return store.SaveConfig(cfg)
}

func AddCustomPortAndSave(state *State, port int) error {
	if err := state.AddCustomPort(port); err != nil {
		return err
	}
	return SaveConfig(state)
}

func RemoveCustomPortAndSave(state *State, port int) error {
	if err := state.RemoveCustomPort(port); err != nil {
		return err
	}
	return SaveConfig(state)
}

func TogglePresetAndSave(state *State, port int, enabled bool) error {
	state.TogglePreset(port, enabled)
	return SaveConfig(state)
}

func TogglePinAndSave(state *State, port int, pinned bool) error {
	state.TogglePin(port, pinned)
	return SaveConfig(state)
}

func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return errors.New("port must be 1-65535")
	}
	return nil
}
