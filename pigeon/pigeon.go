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
	topic    string
	qos      byte
	callback mqtt.MessageHandler
}

func NewPigeon(mqtt *mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) *Pigeon {
	return &Pigeon{
		mqtt:     mqtt,
		topic:    topic,
		qos:      qos,
		callback: callback,
	}
}

// Subscribe to topic
func (p *Pigeon) Subscribe() {
	(*p.mqtt).Subscribe(p.topic, p.qos, p.callback)
}

// Unsubscribe from topic
func (p *Pigeon) Unsubscribe() {
	(*p.mqtt).Unsubscribe(p.topic)
}

type Flock struct {
	mqtt struct {
		opts   *mqtt.ClientOptions
		client *mqtt.Client
	}
	active map[string][]*Pigeon
}

func NewFlock() *Flock {
	f := new(Flock)
	// MQTT
	f.startMQTT()

	// Pigeons
	active := make(map[string][]*Pigeon)
	for _, site := range cfg.Sites {
		var pigeons []*Pigeon
		for _, topic := range site.Topics() {
			pigeon := NewPigeon(f.mqtt.client, topic, 1, nil)
			// Subscribe
			pigeon.Subscribe()
			pigeons = append(pigeons, pigeon)
		}
		active[site.Name] = pigeons
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
				// f.audit()
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
