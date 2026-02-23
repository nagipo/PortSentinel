//go:build darwin || linux

package ports

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"port_sentinel/internal/util"
)

func ScanPort(port int) (PortScanResult, error) {
	results, err := ScanPorts([]int{port})
	if err != nil {
		return PortScanResult{}, err
	}
	if len(results) == 0 {
		return PortScanResult{}, errors.New("no results")
	}
	return results[0], nil
}

func ScanPorts(ports []int) ([]PortScanResult, error) {
	results := make([]PortScanResult, 0, len(ports))
	infoMap, scanErr := scanListeningPorts()
	for _, port := range ports {
		res := PortScanResult{
			Port:      port,
			Status:    StatusFree,
			Protocol:  ProtocolTCP,
			UpdatedAt: NowStamp(),
		}
		if info, ok := infoMap[port]; ok {
			res.Status = StatusInUse
			res.PID = info.PID
			res.LocalAddress = info.LocalAddress
			if info.PID > 0 {
				if pinfo, err := GetProcessInfo(info.PID); err == nil {
					res.ProcessName = pinfo.ProcessName
					res.CommandLine = pinfo.CommandLine
					res.ExePath = pinfo.ExePath
				} else {
					res.Error = err.Error()
				}
			}
		} else if scanErr != nil {
			res.Status = StatusUnknown
			res.Error = scanErr.Error()
		}
		results = append(results, res)
	}
	return results, scanErr
}

func scanListeningPorts() (map[int]PortInfo, error) {
	lsof := util.RunCommand(5*time.Second, "lsof", "-nP", "-iTCP", "-sTCP:LISTEN")
	if lsof.Err == nil {
		return parseLsof(util.CleanOutput(lsof.Stdout)), nil
	}

	netstat := util.RunCommand(5*time.Second, "netstat", "-lntp")
	if netstat.Err == nil {
		return parseUnixNetstat(util.CleanOutput(netstat.Stdout)), nil
	}

	return map[int]PortInfo{}, fmt.Errorf("lsof error: %v; netstat error: %v", lsof.Err, netstat.Err)
}

func GetProcessInfo(pid int) (ProcessInfo, error) {
	if pid <= 0 {
		return ProcessInfo{}, errors.New("invalid pid")
	}
	info := ProcessInfo{PID: pid}

	if runtime.GOOS == "linux" {
		cmdlinePath := filepath.Join("/proc", strconv.Itoa(pid), "cmdline")
		if data, err := os.ReadFile(cmdlinePath); err == nil {
			parts := strings.Split(string(data), "\x00")
			info.CommandLine = strings.TrimSpace(strings.Join(parts, " "))
		}
		exePath := filepath.Join("/proc", strconv.Itoa(pid), "exe")
		if path, err := os.Readlink(exePath); err == nil {
			info.ExePath = path
			info.ProcessName = filepath.Base(path)
		}
	}

	psComm := util.RunCommand(4*time.Second, "ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	if psComm.Err == nil {
		info.ProcessName = strings.TrimSpace(psComm.Stdout)
	}

	psCmd := util.RunCommand(4*time.Second, "ps", "-p", strconv.Itoa(pid), "-o", "command=")
	if psCmd.Err == nil && strings.TrimSpace(psCmd.Stdout) != "" {
		info.CommandLine = strings.TrimSpace(psCmd.Stdout)
	}

	return info, nil
}

func KillPID(pid int, force bool) error {
	if pid <= 0 {
		return errors.New("invalid pid")
	}
	signal := "-15"
	if force {
		signal = "-9"
	}
	res := util.RunCommand(5*time.Second, "kill", signal, strconv.Itoa(pid))
	if res.Err != nil {
		return res.Err
	}
	return nil
}
