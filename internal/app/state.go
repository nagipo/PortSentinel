package app

import (
	"errors"
	"sort"
	"sync"

	"port_sentinel/internal/ports"
	"port_sentinel/internal/store"
)

type State struct {
	mu      sync.Mutex
	Config  store.Config
	Ports   []int
	Results map[int]ports.PortScanResult
}

func NewState(cfg store.Config) *State {
	s := &State{
		Config:  cfg,
		Results: map[int]ports.PortScanResult{},
	}
	s.RebuildPorts()
	return s
}

func (s *State) RebuildPorts() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Ports = buildPortsList(s.Config)
}

func (s *State) UpdateConfig(cfg store.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config = cfg
	s.Ports = buildPortsList(cfg)
}

func (s *State) SnapshotConfig() store.Config {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneConfig(s.Config)
}

func (s *State) GetPorts() []int {
	s.mu.Lock()
	defer s.mu.Unlock()
	portsCopy := make([]int, 0, len(s.Ports))
	portsCopy = append(portsCopy, s.Ports...)
	return portsCopy
}

func (s *State) SetResult(result ports.PortScanResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Results == nil {
		s.Results = map[int]ports.PortScanResult{}
	}
	s.Results[result.Port] = result
}

func (s *State) SetResults(results []ports.PortScanResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Results == nil {
		s.Results = map[int]ports.PortScanResult{}
	}
	for _, res := range results {
		s.Results[res.Port] = res
	}
}

func (s *State) SnapshotResults() []ports.PortScanResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ports.PortScanResult, 0, len(s.Ports))
	for _, port := range s.Ports {
		if res, ok := s.Results[port]; ok {
			out = append(out, res)
		} else {
			out = append(out, ports.PortScanResult{
				Port:      port,
				Status:    ports.StatusUnknown,
				Protocol:  ports.ProtocolTCP,
				UpdatedAt: ports.NowStamp(),
			})
		}
	}
	return out
}

func (s *State) AddCustomPort(port int) error {
	if port <= 0 || port > 65535 {
		return errors.New("port must be 1-65535")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.Config.CustomPorts {
		if existing == port {
			return errors.New("port already exists")
		}
	}
	s.Config.CustomPorts = append(s.Config.CustomPorts, port)
	s.Ports = buildPortsList(s.Config)
	return nil
}

func (s *State) RemoveCustomPort(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make([]int, 0, len(s.Config.CustomPorts))
	found := false
	for _, existing := range s.Config.CustomPorts {
		if existing == port {
			found = true
			continue
		}
		next = append(next, existing)
	}
	if !found {
		return errors.New("port not found")
	}
	s.Config.CustomPorts = next
	if s.Config.PinnedPorts != nil {
		delete(s.Config.PinnedPorts, port)
	}
	s.Ports = buildPortsList(s.Config)
	return nil
}

func (s *State) TogglePreset(port int, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Config.PresetPorts == nil {
		s.Config.PresetPorts = map[int]bool{}
	}
	s.Config.PresetPorts[port] = enabled
	if !enabled && s.Config.PinnedPorts != nil {
		delete(s.Config.PinnedPorts, port)
	}
	s.Ports = buildPortsList(s.Config)
}

func (s *State) TogglePin(port int, pinned bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Config.PinnedPorts == nil {
		s.Config.PinnedPorts = map[int]bool{}
	}
	s.Config.PinnedPorts[port] = pinned
	s.Ports = buildPortsList(s.Config)
}

func (s *State) IsPinned(port int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Config.PinnedPorts != nil && s.Config.PinnedPorts[port]
}

func buildPortsList(cfg store.Config) []int {
	set := map[int]struct{}{}
	for port, enabled := range cfg.PresetPorts {
		if enabled {
			set[port] = struct{}{}
		}
	}
	for _, port := range cfg.CustomPorts {
		set[port] = struct{}{}
	}
	pinned := make([]int, 0, len(set))
	for port, pinnedVal := range cfg.PinnedPorts {
		if pinnedVal {
			if _, ok := set[port]; ok {
				pinned = append(pinned, port)
			}
		}
	}
	sort.Ints(pinned)

	rest := make([]int, 0, len(set))
	for port := range set {
		if cfg.PinnedPorts != nil && cfg.PinnedPorts[port] {
			continue
		}
		rest = append(rest, port)
	}
	sort.Ints(rest)

	return append(pinned, rest...)
}

func cloneConfig(cfg store.Config) store.Config {
	out := cfg

	if cfg.PresetPorts != nil {
		out.PresetPorts = make(map[int]bool, len(cfg.PresetPorts))
		for k, v := range cfg.PresetPorts {
			out.PresetPorts[k] = v
		}
	} else {
		out.PresetPorts = map[int]bool{}
	}

	if cfg.CustomPorts != nil {
		out.CustomPorts = append([]int(nil), cfg.CustomPorts...)
	} else {
		out.CustomPorts = []int{}
	}

	if cfg.PinnedPorts != nil {
		out.PinnedPorts = make(map[int]bool, len(cfg.PinnedPorts))
		for k, v := range cfg.PinnedPorts {
			out.PinnedPorts[k] = v
		}
	} else {
		out.PinnedPorts = map[int]bool{}
	}

	return out
}
