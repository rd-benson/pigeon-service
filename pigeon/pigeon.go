package pigeon

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Pigeon carries data from an MQTT message to InfluxDB
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

type Flock struct {
	mqtt struct {
		opts   *mqtt.ClientOptions
		client *mqtt.Client
	}
	active map[string]*Pigeon
}

func NewFlock() *Flock {
	f := new(Flock)
	// MQTT
	f.startMQTT()

	// Pigeons
	active := make(map[string]*Pigeon)
	for siteName, topics := range cfg.Map() {
		pigeon := NewPigeon(f.mqtt.client, topics, 1, nil)
		pigeon.Subscribe(pigeon.topics...)
		active[siteName] = pigeon
	}
	f.active = active

	return f
}

func (f *Flock) Serve() {
	// Wait for cfg changes
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
				// f.audit(cfgChange.prevCfg)
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
}

/* // audit handles configuration changes to sites
// audit manages MQTT subscriptions and forwarding data on to InfluxDB
func (f *Flock) audit(prevCfg Config) {
	add, remove := common.MapDiffSlice(prevCfg.Map(), cfg.Map())
	for siteName, topics := range add {
		var pigeons []*Pigeon
		for _, topic := range topics {
			pigeon := NewPigeon(f.mqtt.client, topic, 1, nil)
			pigeon.Subscribe()
			pigeons = append(pigeons, pigeon)
		}
		f.active[siteName] = pigeons
	}
	for siteName, topics := range remove {
		for _, pigeon :=
		f.active[siteName] = pigeons
	}
} */
