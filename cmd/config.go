package cmd

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	cfgLock    sync.Mutex
	cfgTimeout = 5 * time.Second
	ErrBlocked = errors.New("f not called: too many calls to RunOnce")
	runningCfg map[string]interface{}
)

// RunOnce runs f and then blocks any further executions
// If f ran, return nil, else ErrBlocked
func RunOnce(f func(), lock *sync.Mutex) error {
	if lock.TryLock() {
		f()
		return nil
	}
	return ErrBlocked
}

// RunOncePerPeriod runs f and then blocks any further executions within the timeout period
// If f ran, return nil, else ErrBlocked
func RunOncePerPeriod(f func(), lock *sync.Mutex, period time.Duration) error {
	if lock.TryLock() {
		f()
		time.AfterFunc(period, lock.Unlock)
		return nil
	}
	return ErrBlocked
}

// Initialise viper using custom OnConfigChange function that stops multiple ops
// during a short period of time
func initConfig() {
	viper.AddConfigPath("./")
	viper.ReadInConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		RunOncePerPeriod(func() {
			determineChanges()
		}, &cfgLock, cfgTimeout)
	})
	viper.WatchConfig()
	// Save running config for comparison on config change
	runningCfg = viper.AllSettings()
}

// Determine which parts of configuration have changed
func determineChanges() {
	fmt.Println("determineChanges called")
}
