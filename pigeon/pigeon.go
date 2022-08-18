package pigeon

import (
	"fmt"
	"time"

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
		opts   mqtt.ClientOptions
		client mqtt.Client
	}
	active map[string][]*Pigeon
}

func NewFlock() *Flock {
	f := new(Flock)
	// MQTT
	// Options
	f.mqtt.opts = *mqtt.NewClientOptions()
	f.mqtt.opts.AddBroker(cfg.MQTT.URI())
	f.mqtt.opts.SetClientID("pigeon")
	f.mqtt.opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		// fmt.Printf("TOPIC: %v\n", m.Topic())
		// fmt.Printf("PAYLOAD: %v\n", string(m.Payload()))
	})
	// Client
	f.mqtt.client = mqtt.NewClient(&f.mqtt.opts)
	if token := f.mqtt.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// Pigeons
	active := make(map[string][]*Pigeon)
	for _, site := range cfg.Sites {
		var pigeons []*Pigeon
		for _, topic := range site.Topics() {
			pigeon := NewPigeon(&f.mqtt.client, topic, 1, nil)
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
		fmt.Println("pigeon serve called")
		for {
			select {
			case <-cfgChange.mqtt:
				fmt.Println("MQTT config changed!")
			case <-cfgChange.influxdb:
				fmt.Println("Database config changed!")
			case <-cfgChange.sites:
				fmt.Println("Sites config changed!")
			default:
				fmt.Println("Config unchanged!")
				time.Sleep(timeout)
			}
		}
	}()
}
