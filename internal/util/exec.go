package util

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

type CmdResult struct {
	Stdout string
	Stderr string
	Err    error
}

func RunCommand(timeout time.Duration, name string, args ...string) CmdResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return CmdResult{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
			Err:    ctx.Err(),
		}
	}

	return CmdResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
}

func CleanOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimSpace(s)
}
