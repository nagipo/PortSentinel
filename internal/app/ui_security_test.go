//go:build cgo

package app

import (
	"strings"
	"testing"
)

func TestMaskSensitiveArgsMasksKnownSecrets(t *testing.T) {
	input := "server --token=abc123 --password:letmein --api_key =k123 Authorization: Bearer qwerty.zzz"
	out := maskSensitiveArgs(input)

	for _, secret := range []string{"abc123", "letmein", "k123", "qwerty.zzz"} {
		if strings.Contains(out, secret) {
			t.Fatalf("expected secret %q to be masked, got: %s", secret, out)
		}
	}
	if !strings.Contains(out, "token=***") {
		t.Fatalf("expected token to be masked, got: %s", out)
	}
	if !strings.Contains(out, "password=***") {
		t.Fatalf("expected password to be masked, got: %s", out)
	}
	if !strings.Contains(out, "api_key=***") {
		t.Fatalf("expected api_key to be masked, got: %s", out)
	}
	if !strings.Contains(strings.ToLower(out), "bearer ***") {
		t.Fatalf("expected bearer token to be masked, got: %s", out)
	}
}

func TestMaskSensitiveArgsKeepsEmptyInput(t *testing.T) {
	if got := maskSensitiveArgs("   "); got != "   " {
		t.Fatalf("expected whitespace input unchanged, got %q", got)
	}
}
