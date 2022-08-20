package pigeon

import (
	"fmt"
	"reflect"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	v         = viper.New()
	cfg       = new(Config)
	cfgChange = OnChangeChan{
		make(chan bool, 1),
		make(chan bool, 1),
		make(chan bool, 1),
		Config{},
	}
	allowCfgChange   = make(chan bool, 1)
	cfgChangeLimiter *time.Ticker
)

func init() {
	// fsnotify fires twice when edited with rich text editor
	// below allows only one call per second
	allowCfgChange <- true
	cfgChangeLimiter = time.NewTicker(1 * time.Second)
	go func() {
		for range cfgChangeLimiter.C {
			allowCfgChange <- true
		}
	}()
}

// Main pigeon configuration
type Config struct {
	MQTT     MQTT     `validate:"required"`
	InfluxDB InfluxDB `validate:"required"`
	Sites    []Site   `validate:"required"`
}

// SiteMap returns a map of site names to device topics
func (c *Config) Map() map[string][]string {
	m := make(map[string][]string)
	for _, site := range c.Sites {
		m[site.Name] = site.Topics()
	}
	return m
}

// Initialise viper, watch for changes and send signal to channel
// specifiying which aspect of channel changed via OnConfigChange
func WatchConfig(path string) {
	v.AddConfigPath(path)
	v.ReadInConfig()
	// If config errors during new, panic!
	if err := Unmarshal(cfg); err != nil {
		panic(err)
	}

	v.OnConfigChange(func(e fsnotify.Event) {
		v.ReadInConfig()
		// Occasionally viper fails to read
		if len(v.AllSettings()) == 0 {
			return
		}

		select {
		case <-allowCfgChange:
			OnConfigChange()
		default:
			return
		}
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

// OnChangeChan contains channels to notify the flock of a config change
type OnChangeChan struct {
	mqtt     chan bool
	influxdb chan bool
	sites    chan bool
	prevCfg  Config
}

// Determine which parts of configuration have changed
// and signal flock via OnChangeChan
func OnConfigChange() {
	tmpCfg := Config{}
	prevCfg := *cfg
	if err := Unmarshal(&tmpCfg); err == nil {
		*cfg = tmpCfg
	} else {
		fmt.Println("pigeon will not update due to errors in config: ", err)
		return
	}
	// Check for configuration changes and send signal to flock
	if !reflect.DeepEqual(prevCfg.MQTT, (*cfg).MQTT) {
		cfgChange.mqtt <- true
	}
	if !reflect.DeepEqual(prevCfg.InfluxDB, (*cfg).InfluxDB) {
		cfgChange.influxdb <- true
	}
	if !reflect.DeepEqual(prevCfg.Sites, (*cfg).Sites) {
		cfgChange.sites <- true
	}
	cfgChange.prevCfg = prevCfg
}

// MQTT configuration
type MQTT struct {
	FQDN string `validate:"required"`
	Port uint16 `validate:"required"`
}

// URI returns MQTT broker URI
func (m *MQTT) URI() string {
	return fmt.Sprintf("tcp://%v:%v", m.FQDN, m.Port)
}

// InfluxDB configuration
type InfluxDB struct {
	FQDN string `validate:"required"`
	// TokenRead  string `validate:"required"`
	TokenWrite string `validate:"required"`
	OrgName    string `validate:"email,required"`
	OrgId      string `validate:"required"`
}

func (i *InfluxDB) URI() string {
	return fmt.Sprintf("https://%v", i.FQDN)
}

// Site configuration
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
