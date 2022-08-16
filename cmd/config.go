package cmd

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	cfgLock    sync.Mutex
	cfgTimeout = 1 * time.Second
	ErrBlocked = errors.New("f not called: too many calls to RunOnce")
	runningCfg Config
	tmpCfg     Config
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
			viper.ReadInConfig()
			updateRunningConfig()
		}, &cfgLock, cfgTimeout)
	})
	viper.WatchConfig()
	// Save running config for comparison on config change
	// If config errors during init, panic!
	if err := Unmarshal(&tmpCfg); err != nil {
		panic(err)
	}
	runningCfg = tmpCfg
}

// Determine which parts of configuration have changed
func updateRunningConfig() {
	tmpCfg = Config{}
	prevCfg := runningCfg
	if err := Unmarshal(&tmpCfg); err == nil {
		runningCfg = tmpCfg
	} else {
		fmt.Println("pigeon will not update due to errors in config: ", err)
		return
	}
	// Check if broker config changed
	if !reflect.DeepEqual(prevCfg.MQTT, runningCfg.MQTT) {
		restartMqttClient()
	}
	// Check if database config changed
	if !reflect.DeepEqual(prevCfg.InfluxDB, runningCfg.InfluxDB) {
		// TODO restart database client
		fmt.Println("database config changed")
	}
	// Check if sites config changed
	if !reflect.DeepEqual(prevCfg.Sites, runningCfg.Sites) {
		// TODO subscribe/unsubscribe to topics
		// TODO remove connection to database
		fmt.Println("sites config changed")
	}

}

// Unmarshal viper configuration with validation checks
// pigeon won't start unless the configuration is valid
func Unmarshal(c interface{}) error {
	viper.Unmarshal(c)
	err := validator.New().Struct(c)
	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		if len(validationErrors) > 0 {
			if err != nil {
				return errors.Wrap(err, "validation error")
			}
		}
		return nil
	}
	return nil
}

type Config struct {
	MQTT     MQTT     `validate:"required"`
	InfluxDB InfluxDB `validate:"required"`
	Sites    []Site   `validate:"required"`
}

type MQTT struct {
	FQDN string `validate:"fqdn"`
	Port uint16 `validate:"numeric"`
}

type InfluxDB struct {
	FQDN         string `validate:"fqdn"`
	TokenRead    string `validate:"required"`
	TokenWrite   string `validate:"required"`
	Organisation string `validate:"email"`
}

type Site struct {
	Name    string
	Devices []string
}
