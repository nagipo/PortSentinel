package ports

import (
	"strconv"
	"strings"
)

func parseWindowsNetstat(output string) map[int]PortInfo {
	out := map[int]PortInfo{}
	lines := strings.Split(output, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if !strings.EqualFold(fields[0], "TCP") {
			continue
		}
		state := strings.ToUpper(fields[3])
		if state != "LISTENING" {
			continue
		}
		port := parsePortFromAddress(fields[1])
		if port == 0 {
			continue
		}
		pid, _ := strconv.Atoi(fields[4])
		out[port] = PortInfo{
			PID:          pid,
			LocalAddress: fields[1],
		}
	}
	return out
}

func parseLsof(output string) map[int]PortInfo {
	out := map[int]PortInfo{}
	lines := strings.Split(output, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "COMMAND") {
			continue
		}
		if !strings.Contains(line, "(LISTEN)") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pid, _ := strconv.Atoi(fields[1])
		addr := extractLsofAddress(line)
		port := parsePortFromAddress(addr)
		if port == 0 {
			continue
		}
		out[port] = PortInfo{
			PID:          pid,
			LocalAddress: addr,
		}
	}
	return out
}

func extractLsofAddress(line string) string {
	// Typical line ends with: "TCP *:3000 (LISTEN)" or "TCP 127.0.0.1:8080 (LISTEN)"
	start := strings.Index(line, "TCP ")
	if start == -1 {
		return ""
	}
	segment := line[start+4:]
	end := strings.Index(segment, " (LISTEN)")
	if end != -1 {
		segment = segment[:end]
	}
	segment = strings.TrimSpace(segment)
	parts := strings.Fields(segment)
	if len(parts) == 0 {
		return ""
	}
	// Address usually the last token in the segment.
	return parts[len(parts)-1]
}

func parseUnixNetstat(output string) map[int]PortInfo {
	out := map[int]PortInfo{}
	lines := strings.Split(output, "\n")
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "LISTEN") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		local := fields[3]
		port := parsePortFromAddress(local)
		if port == 0 {
			continue
		}
		last := fields[len(fields)-1]
		pid := 0
		if idx := strings.Index(last, "/"); idx > 0 {
			pid, _ = strconv.Atoi(last[:idx])
		}
		out[port] = PortInfo{
			PID:          pid,
			LocalAddress: local,
		}
	}
	return out
}

func parsePortFromAddress(addr string) int {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return 0
	}
	addr = strings.TrimPrefix(addr, "[")
	addr = strings.TrimSuffix(addr, "]")
	if strings.Contains(addr, "]") {
		addr = strings.TrimPrefix(addr, "[")
		addr = strings.TrimSuffix(addr, "]")
	}
	if strings.Contains(addr, "*:") {
		parts := strings.Split(addr, ":")
		if len(parts) > 0 {
			addr = parts[len(parts)-1]
		}
	}
	lastColon := strings.LastIndex(addr, ":")
	if lastColon >= 0 {
		addr = addr[lastColon+1:]
	}
	port, _ := strconv.Atoi(strings.TrimSpace(addr))
	return port
}
