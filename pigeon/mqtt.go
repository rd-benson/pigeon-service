package pigeon

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// InitMqtt returns MQTT options, client and error
// Error is nil on success
func InitMqtt() (*mqtt.ClientOptions, *mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.URI())
	opts.SetClientID("pigeon")
	opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		// fmt.Printf("TOPIC: %v\n", m.Topic())
		// fmt.Printf("PAYLOAD: %v\n", string(m.Payload()))
	})
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return opts, &client, token.Error()
	}
	return opts, &client, nil
}
