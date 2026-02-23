package app

import (
	"errors"
	"os"

	"port_sentinel/internal/ports"
	"port_sentinel/internal/store"
)

type PortScanner interface {
	ScanPorts(ports []int) ([]ports.PortScanResult, error)
	ScanPort(port int) (ports.PortScanResult, error)
	KillPID(pid int, force bool) error
}

type ConfigRepository interface {
	SaveConfig(cfg store.Config) error
}

type Service struct {
	state   *State
	scanner PortScanner
	repo    ConfigRepository
}

func NewService(state *State, scanner PortScanner, repo ConfigRepository) *Service {
	return &Service{
		state:   state,
		scanner: scanner,
		repo:    repo,
	}
}

func NewDefaultService(state *State) *Service {
	return NewService(state, osPortScanner{}, fileConfigRepository{})
}

func (s *Service) RefreshAll() ([]ports.PortScanResult, error) {
	portsList := s.state.GetPorts()
	results, err := s.scanner.ScanPorts(portsList)
	if len(results) > 0 {
		s.state.SetResults(results)
	}
	return results, err
}

func (s *Service) RefreshOne(port int) (ports.PortScanResult, error) {
	res, err := s.scanner.ScanPort(port)
	if err == nil {
		s.state.SetResult(res)
	}
	return res, err
}

func (s *Service) KillProcess(pid int, force bool) error {
	if pid == os.Getpid() {
		return errors.New("refusing to terminate Port Sentinel itself")
	}
	return s.scanner.KillPID(pid, force)
}

func (s *Service) SaveConfig() error {
	cfg := s.state.SnapshotConfig()
	return s.repo.SaveConfig(cfg)
}

func (s *Service) UpdateUIConfig(updater func(cfg *store.Config) error) error {
	cfg := s.state.SnapshotConfig()
	if err := updater(&cfg); err != nil {
		return err
	}
	s.state.UpdateConfig(cfg)
	return s.repo.SaveConfig(cfg)
}

func (s *Service) AddCustomPortAndSave(port int) error {
	if err := s.state.AddCustomPort(port); err != nil {
		return err
	}
	return s.SaveConfig()
}

func (s *Service) RemoveCustomPortAndSave(port int) error {
	if err := s.state.RemoveCustomPort(port); err != nil {
		return err
	}
	return s.SaveConfig()
}

func (s *Service) TogglePresetAndSave(port int, enabled bool) error {
	s.state.TogglePreset(port, enabled)
	return s.SaveConfig()
}

func (s *Service) TogglePinAndSave(port int, pinned bool) error {
	s.state.TogglePin(port, pinned)
	return s.SaveConfig()
}

func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return errors.New("port must be 1-65535")
	}
	return nil
}

type osPortScanner struct{}

func (osPortScanner) ScanPorts(portList []int) ([]ports.PortScanResult, error) {
	return ports.ScanPorts(portList)
}

func (osPortScanner) ScanPort(port int) (ports.PortScanResult, error) {
	return ports.ScanPort(port)
}

func (osPortScanner) KillPID(pid int, force bool) error {
	return ports.KillPID(pid, force)
}

type fileConfigRepository struct{}

func (fileConfigRepository) SaveConfig(cfg store.Config) error {
	return store.SaveConfig(cfg)
}
