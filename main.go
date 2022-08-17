package main

import (
	"time"

	"github.com/rd-benson/pigeon-service/pigeon"
)

func main() {
	pigeon.WatchConfig("./")
	for {
		time.Sleep(1 * time.Second)
	}
}
