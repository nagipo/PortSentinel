//go:build !cgo

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Port Sentinel requires CGO-enabled Go toolchain to build the Fyne UI. Set CGO_ENABLED=1 and ensure a working C compiler is installed.")
	os.Exit(1)
}
