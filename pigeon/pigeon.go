package pigeon

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rd-benson/pigeon-service/common"
)

// Pigeon carries data from an MQTT message to InfluxDB: one pigeon per site
// Pigeon struct contains all necessary information to:
// - get data from MQTT broker (topic, qos)
// TODO
// - pass data to InfluxDB (callback, bucket)
type Pigeon struct {
	mqtt     *mqtt.Client
	topics   []string
	qos      byte
	callback mqtt.MessageHandler
}

func NewPigeon(mqtt *mqtt.Client, topics []string, qos byte, callback mqtt.MessageHandler) *Pigeon {
	return &Pigeon{
		mqtt:     mqtt,
		topics:   topics,
		qos:      qos,
		callback: callback,
	}
}

// Subscribe to topic(s)
func (p *Pigeon) Subscribe(topics ...string) {
	for _, topic := range topics {
		(*p.mqtt).Subscribe(topic, p.qos, p.callback)
	}
}

// Unsubscribe from topic(s)
func (p *Pigeon) Unsubscribe(topics ...string) {
	for _, topic := range topics {
		(*p.mqtt).Unsubscribe(topic)
	}
}

// Flock manages all pigeons
type Flock struct {
	mqtt struct {
		opts   *mqtt.ClientOptions
		client *mqtt.Client
	}
	influxdb struct {
		client  *influxdb2.Client
		writers map[string]*api.WriteAPI
	}
	active map[string]*Pigeon
}

func NewFlock() *Flock {
	f := new(Flock)
	// MQTT
	f.startMQTT()

	// InfluxDB
	f.startInfluxDB()

	// Pigeons
	f.active = map[string]*Pigeon{}
	f.audit(Config{})

	return f
}

// Watch for config changes and command pigeons accordingly
func (f *Flock) Watch() {
	go func() {
		for {
			select {
			case <-cfgChange.mqtt:
				// TODO could instead try to create new client connection
				// on failure prompt user (catch typos etc.)
				// For now, just reinit client (and don't make mistakes in config...)
				f.startMQTT()
			case <-cfgChange.influxdb:
				// TODO restart database client
				fmt.Println("Database config changed!")
			case <-cfgChange.sites:
				// Subscribe/unsubscribe to topics
				f.audit(cfgChange.prevCfg)
				// TODO remove connection to database
			}
		}
	}()
}

// startMQTT links flock to an MQTT client
// If the connection to the client is bad, pigeon will exit with status code 1
func (f *Flock) startMQTT() {
	var err error
	f.mqtt.opts, f.mqtt.client, err = NewMQTT()
	if err != nil {
		fmt.Println("pigeon cannot continue: ", err)
		os.Exit(1)
	}
	fmt.Println("MQTT ok")
}

// startInfluxDB links flock to an influxdb client and initialises one write API per site
// If the connection to the client is bad, pigeon will exit with status code 2
func (f *Flock) startInfluxDB() {
	// Get new client, will exit with status code 2 if write token is invalid
	f.influxdb.client = NewInfluxDB()
	// Initialise writers
	f.influxdb.writers = map[string]*api.WriteAPI{}
	for _, site := range cfg.Sites {
		created := CreateBucketSafe(site.Name, f.influxdb.client)
		if created {
			fmt.Printf("created bucket: %v\n", site.Name)
		}
		writer := (*f.influxdb.client).WriteAPI(cfg.InfluxDB.OrgName, site.Name)
		f.influxdb.writers[site.Name] = &writer
	}
	fmt.Println("InfluxDB ok")
}

// audit handles configuration changes to sites
// audit manages MQTT subscriptions and forwarding data on to InfluxDB
func (f *Flock) audit(prevCfg Config) {
	add, remove := common.MapDiffSlice(prevCfg.Map(), cfg.Map())
	for site, topics := range add {
		pigeon := NewPigeon(f.mqtt.client, topics, 1, nil)
		pigeon.Subscribe(pigeon.topics...)
		f.active[site] = pigeon
	}
	for site, topics := range remove {
		f.active[site].Unsubscribe(topics...)
	}
}
