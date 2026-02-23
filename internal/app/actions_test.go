package app

import (
	"os"
	"strings"
	"testing"

	"port_sentinel/internal/ports"
	"port_sentinel/internal/store"
)

type fakeScanner struct {
	killedPID   int
	killedForce bool
	killErr     error
}

func (f *fakeScanner) ScanPorts(_ []int) ([]ports.PortScanResult, error) {
	return nil, nil
}

func (f *fakeScanner) ScanPort(_ int) (ports.PortScanResult, error) {
	return ports.PortScanResult{}, nil
}

func (f *fakeScanner) KillPID(pid int, force bool) error {
	f.killedPID = pid
	f.killedForce = force
	return f.killErr
}

type fakeRepo struct{}

func (fakeRepo) SaveConfig(_ store.Config) error {
	return nil
}

func TestServiceKillProcessRejectsSelf(t *testing.T) {
	state := NewState(store.DefaultConfig())
	scanner := &fakeScanner{}
	svc := NewService(state, scanner, fakeRepo{})

	err := svc.KillProcess(os.Getpid(), true)
	if err == nil {
		t.Fatalf("expected self-kill to be rejected")
	}
	if !strings.Contains(err.Error(), "refusing") {
		t.Fatalf("expected refusing error, got: %v", err)
	}
	if scanner.killedPID != 0 {
		t.Fatalf("expected scanner not to be called, got pid=%d", scanner.killedPID)
	}
}

func TestServiceKillProcessDelegatesToScanner(t *testing.T) {
	state := NewState(store.DefaultConfig())
	scanner := &fakeScanner{}
	svc := NewService(state, scanner, fakeRepo{})

	targetPID := os.Getpid() + 1000
	if err := svc.KillProcess(targetPID, true); err != nil {
		t.Fatalf("expected delegate kill to pass, got: %v", err)
	}
	if scanner.killedPID != targetPID {
		t.Fatalf("expected killed pid %d, got %d", targetPID, scanner.killedPID)
	}
	if !scanner.killedForce {
		t.Fatalf("expected force flag to be true")
	}
}
