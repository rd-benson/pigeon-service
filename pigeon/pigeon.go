package pigeon

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Pigeon is the interface definition to allow mocking tests.
type Pigeon interface {
	Subscribe()
	Unsubscribe()
}

// pigeon carries data from an MQTT message to InfluxDB
// pigeon struct contains all necessary information to:
// - get data from MQTT broker (topic, qos)
// TODO
// - pass data to InfluxDB (callback, bucket)
type pigeon struct {
	mqtt     *mqtt.Client
	topic    string
	qos      byte
	callback mqtt.MessageHandler
}

func NewPigeon(mqtt *mqtt.Client, topic string, qos byte, callback mqtt.MessageHandler) Pigeon {
	return &pigeon{
		mqtt:     mqtt,
		topic:    topic,
		qos:      qos,
		callback: callback,
	}
}

// Subscribe to topic
func (p *pigeon) Subscribe() {
	(*p.mqtt).Subscribe(p.topic, p.qos, p.callback)
}

// Unsubscribe from topic
func (p *pigeon) Unsubscribe() {
	(*p.mqtt).Unsubscribe(p.topic)
}
