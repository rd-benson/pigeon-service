package main

import (
	"time"

	"github.com/rd-benson/pigeon-service/cmd"
)

func main() {
	cmd.Start()

	// Serve forever
	for {
		time.Sleep(1 * time.Second)
	}
}
