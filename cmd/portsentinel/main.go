//go:build cgo

package main

import (
	"log"

	"port_sentinel/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
