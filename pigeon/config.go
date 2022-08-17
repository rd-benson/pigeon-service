package pigeon

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
	v          = viper.New()
	cfg        = new(Config)
	lock       = new(sync.Mutex)
	timeout    = 500 * time.Millisecond
	ErrBlocked = errors.New("f not called: too many calls to RunOnce")
)

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

type Config struct {
	MQTT     MQTT     `validate:"required"`
	InfluxDB InfluxDB `validate:"required"`
	Sites    []Site   `validate:"required"`
}

// Initialise viper using custom OnConfigChange function that stops multiple ops
// during a short period of time
func WatchConfig(path string) {
	v.AddConfigPath(path)
	v.ReadInConfig()
	// If config errors during new, panic!
	if err := Unmarshal(cfg); err != nil {
		panic(err)
	}
	v.OnConfigChange(func(e fsnotify.Event) {
		RunOncePerPeriod(func() {
			v.ReadInConfig()
			// Occasionally viper fails to read
			if len(v.AllSettings()) == 0 {
				return
			}
			OnConfigChange()
		}, lock, timeout)
	})
	v.WatchConfig()
}

// Unmarshal viper configuration with validation checks
// pigeon won't start unless the configuration is valid
func Unmarshal(c interface{}) error {
	v.Unmarshal(c)
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

// Determine which parts of configuration have changed
func OnConfigChange() {
	tmpCfg := Config{}
	prevCfg := *cfg
	if err := Unmarshal(&tmpCfg); err == nil {
		*cfg = tmpCfg
	} else {
		fmt.Println("pigeon will not update due to errors in config: ", err)
		return
	}
	// Check if broker config changed
	if !reflect.DeepEqual(prevCfg.MQTT, (*cfg).MQTT) {
		fmt.Println("broker config changed")
		// restartMQTT()
	}
	// Check if database config changed
	if !reflect.DeepEqual(prevCfg.InfluxDB, (*cfg).InfluxDB) {
		// TODO restart database client
		fmt.Println("database config changed")
	}
	// Check if sites config changed
	if !reflect.DeepEqual(prevCfg.Sites, (*cfg).Sites) {
		// TODO subscribe/unsubscribe to topics
		// TODO remove connection to database
		fmt.Println("sites config changed")
	}

}

type MQTT struct {
	FQDN string `validate:"required"`
	Port uint16 `validate:"required"`
}

// URI returns MQTT broker URI
func (m *MQTT) URI() string {
	return fmt.Sprintf("tcp://%v:%v", m.FQDN, m.Port)
}

type InfluxDB struct {
	FQDN         string `validate:"required"`
	TokenRead    string `validate:"required"`
	TokenWrite   string `validate:"required"`
	Organisation string `validate:"email"`
}

type Site struct {
	Name    string
	Devices []string
}

// Topic returns MQTT topic from site
// Easylog should be configured to "publish devices separately"
func (s *Site) Topics() []string {
	t := []string{}
	for _, device := range s.Devices {
		t = append(t, fmt.Sprintf("%v/%v", s.Name, device))
	}
	return t
}
