package pigeon

/* import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqttClient        mqtt.Client
	mqttClientOptions = mqtt.NewClientOptions()
)

// Start connection to MQTT broker given in configuration
func startMQTT(cfg *Config) error {
	fmt.Println("starting MQTT client ...")
	fmt.Printf("connecting to: %v\n", cfg.MQTT.URI())
	// Options
	mqttClientOptions.AddBroker(runningCfg.MQTT.URI())
	mqttClientOptions.SetClientID("pigeon")
	mqttClientOptions.SetOnConnectHandler(func(c mqtt.Client) {
		fmt.Println("... connected!")
	})
	// Client proper
	mqttClient = mqtt.NewClient(mqttClientOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func restartMQTT() {
	fmt.Println("broker config changed: restarting MQTT client")
	// Close existing connection
	mqttClient.Disconnect(500)
	// Start fresh connection
	startMQTT()
} */
