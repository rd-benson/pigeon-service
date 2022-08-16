package cmd

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqttClient        mqtt.Client
	mqttClientOptions = mqtt.NewClientOptions()
)

// Start connection to MQTT broker given in configuration
func startMQTT() {
	fmt.Println("starting MQTT client ...")
	fmt.Printf("connecting to: %v\n", runningCfg.MQTT.URI())
	// Options
	mqttClientOptions.AddBroker(runningCfg.MQTT.URI())
	mqttClientOptions.SetClientID("pigeon")
	mqttClientOptions.SetOnConnectHandler(func(c mqtt.Client) {
		fmt.Println("... connected!")
	})
	// Client proper
	mqttClient = mqtt.NewClient(mqttClientOptions)
	mqttClient.Connect()
}

func restartMQTT() {
	fmt.Println("broker config changed: restarting MQTT client")
	// Close existing connection
	mqttClient.Disconnect(500)
	// Start fresh connection
	startMQTT()
}
