package ports

import (
	"sort"
	"time"
)

type PortStatus string

const (
	StatusFree    PortStatus = "FREE"
	StatusInUse   PortStatus = "IN_USE"
	StatusUnknown PortStatus = "UNKNOWN"
)

type Protocol string

const (
	ProtocolTCP     Protocol = "tcp"
	ProtocolUDP     Protocol = "udp"
	ProtocolUnknown Protocol = "unknown"
)

type PortScanResult struct {
	Port         int        `json:"port"`
	Status       PortStatus `json:"status"`
	Protocol     Protocol   `json:"protocol"`
	PID          int        `json:"pid"`
	ProcessName  string     `json:"processName"`
	CommandLine  string     `json:"commandLine"`
	ExePath      string     `json:"exePath"`
	LocalAddress string     `json:"localAddress"`
	Error        string     `json:"error"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type ProcessInfo struct {
	PID         int
	ProcessName string
	CommandLine string
	ExePath     string
}

type PortInfo struct {
	PID          int
	LocalAddress string
}

func DefaultPresetPorts() map[int]bool {
	return map[int]bool{
		3000:  true,
		5173:  true,
		8080:  true,
		8000:  true,
		5000:  true,
		4200:  true,
		5432:  true,
		6379:  true,
		27017: true,
		9229:  true,
		15672: true,
		5672:  true,
		3306:  true,
		11211: true,
	}
}

func SortedPorts(ports []int) []int {
	cp := make([]int, 0, len(ports))
	cp = append(cp, ports...)
	sort.Ints(cp)
	return cp
}

func NowStamp() time.Time {
	return time.Now().UTC()
}
