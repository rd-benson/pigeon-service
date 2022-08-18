package pigeon

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func InitMqtt() (*mqtt.ClientOptions, *mqtt.Client) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTT.URI())
	opts.SetClientID("pigeon")
	opts.SetDefaultPublishHandler(func(c mqtt.Client, m mqtt.Message) {
		// fmt.Printf("TOPIC: %v\n", m.Topic())
		// fmt.Printf("PAYLOAD: %v\n", string(m.Payload()))
	})
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return opts, nil
	}
	return opts, &client
}
