//go:build windows

package ports

import (
	"errors"
	"fmt"
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

	netstat := util.RunCommand(5*time.Second, "netstat", "-ano", "-p", "tcp")
	if netstat.Err != nil {
		for _, port := range ports {
			results = append(results, PortScanResult{
				Port:      port,
				Status:    StatusUnknown,
				Protocol:  ProtocolTCP,
				Error:     netstat.Err.Error(),
				UpdatedAt: NowStamp(),
			})
		}
		return results, netstat.Err
	}
	infoMap := parseWindowsNetstat(util.CleanOutput(netstat.Stdout))
	procInfoCache := map[int]ProcessInfo{}
	procErrCache := map[int]string{}

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
				if pinfo, ok := procInfoCache[info.PID]; ok {
					res.ProcessName = pinfo.ProcessName
					res.CommandLine = pinfo.CommandLine
					res.ExePath = pinfo.ExePath
				} else if errMsg, ok := procErrCache[info.PID]; ok {
					res.Error = errMsg
				} else if pinfo, err := GetProcessInfo(info.PID); err == nil {
					procInfoCache[info.PID] = pinfo
					res.ProcessName = pinfo.ProcessName
					res.CommandLine = pinfo.CommandLine
					res.ExePath = pinfo.ExePath
				} else {
					procErrCache[info.PID] = err.Error()
					res.Error = err.Error()
				}
			}
		}
		results = append(results, res)
	}

	return results, nil
}

func GetProcessInfo(pid int) (ProcessInfo, error) {
	if pid <= 0 {
		return ProcessInfo{}, errors.New("invalid pid")
	}
	info := ProcessInfo{PID: pid}

	tasklist := util.RunCommand(5*time.Second, "tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
	if tasklist.Err != nil {
		return info, tasklist.Err
	}
	line := util.CleanOutput(tasklist.Stdout)
	if line != "" && strings.Contains(line, "\",\"") {
		parts := splitCSVLine(line)
		if len(parts) > 0 {
			info.ProcessName = parts[0]
		}
	}

	wmic := util.RunCommand(6*time.Second, "wmic", "process", "where", fmt.Sprintf("processid=%d", pid), "get", "CommandLine,ExecutablePath", "/FORMAT:LIST")
	if wmic.Err == nil {
		for _, raw := range strings.Split(util.CleanOutput(wmic.Stdout), "\n") {
			line := strings.TrimSpace(raw)
			if strings.HasPrefix(line, "CommandLine=") {
				info.CommandLine = strings.TrimPrefix(line, "CommandLine=")
			}
			if strings.HasPrefix(line, "ExecutablePath=") {
				info.ExePath = strings.TrimPrefix(line, "ExecutablePath=")
			}
		}
	}

	return info, nil
}

func KillPID(pid int, force bool) error {
	if pid <= 0 {
		return errors.New("invalid pid")
	}
	args := []string{"/PID", strconv.Itoa(pid), "/T"}
	if force {
		args = append(args, "/F")
	}
	res := util.RunCommand(8*time.Second, "taskkill", args...)
	if res.Err != nil {
		return res.Err
	}
	return nil
}

func splitCSVLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	line = strings.TrimPrefix(line, "\"")
	line = strings.TrimSuffix(line, "\"")
	return strings.Split(line, "\",\"")
}
