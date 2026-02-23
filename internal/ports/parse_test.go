package ports

import "testing"

func TestParseWindowsNetstat(t *testing.T) {
	sample := `
  TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       708
  TCP    127.0.0.1:3000         0.0.0.0:0              LISTENING       4242
  TCP    127.0.0.1:3001         0.0.0.0:0              ESTABLISHED     4242
`
	out := parseWindowsNetstat(sample)
	if info, ok := out[3000]; !ok || info.PID != 4242 {
		t.Fatalf("expected port 3000 pid 4242, got %+v", info)
	}
	if _, ok := out[3001]; ok {
		t.Fatalf("expected port 3001 to be excluded")
	}
}

func TestParseLsof(t *testing.T) {
	sample := `
COMMAND   PID USER   FD   TYPE             DEVICE SIZE/OFF NODE NAME
node     1234 user   23u  IPv6 0x1234      0t0  TCP *:3000 (LISTEN)
`
	out := parseLsof(sample)
	if info, ok := out[3000]; !ok || info.PID != 1234 {
		t.Fatalf("expected port 3000 pid 1234, got %+v", info)
	}
}

func TestParseUnixNetstat(t *testing.T) {
	sample := `
tcp        0      0 0.0.0.0:22      0.0.0.0:*     LISTEN      1000/sshd
tcp6       0      0 :::8080         :::*          LISTEN      2000/java
`
	out := parseUnixNetstat(sample)
	if info, ok := out[22]; !ok || info.PID != 1000 {
		t.Fatalf("expected port 22 pid 1000, got %+v", info)
	}
	if info, ok := out[8080]; !ok || info.PID != 2000 {
		t.Fatalf("expected port 8080 pid 2000, got %+v", info)
	}
}
