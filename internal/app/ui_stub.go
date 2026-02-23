//go:build !cgo

package app

import "errors"

func Run() error {
	return errors.New("Fyne UI requires CGO-enabled Go toolchain; rebuild with CGO_ENABLED=1")
}
