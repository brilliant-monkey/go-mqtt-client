package mqtt

import (
	"log"
	"os"
	"time"

	"github.com/brilliant-monkey/go-mqtt-client/types"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	config     types.MQTTConfig
	connection MQTT.Client
	handler    *func(message []byte)
}

func NewMQTTClient(config types.MQTTConfig) *MQTTClient {
	client := &MQTTClient{
		config: config,
	}
	setupClient(client)
	return client
}

func setupClient(mqttClient *MQTTClient) {
	var isReconnecting = false

	clientId, exists := os.LookupEnv("POD_NAME")
	if !exists {
		clientId = mqttClient.config.GetClientId()

		log.Printf("Failed to get POD_NAME. Using local client id %s.", clientId)
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttClient.config.GetEndpoint())
	opts.SetClientID(clientId)
	opts.SetUsername(mqttClient.config.GetUsername())
	opts.SetPassword(mqttClient.config.GetPassword())
	opts.SetAutoReconnect(true)

	opts.OnConnectionLost = func(client MQTT.Client, err error) {
		log.Println("MQTT client disconnected.")
		log.Println(err.Error())
	}
	opts.OnReconnecting = func(client MQTT.Client, opts *MQTT.ClientOptions) {
		log.Println("MQTT reconnecting...")
		isReconnecting = true
	}

	opts.OnConnect = func(client MQTT.Client) {
		log.Println("MQTT connected.")

		if isReconnecting && mqttClient.handler != nil {
			isReconnecting = false
			mqttClient.subscribe(*mqttClient.handler)
		}
	}

	if mqttClient.config.GetProducerTopic() == nil {
		log.Println("MQTT producer config not set.")
	}

	if mqttClient.config.GetConsumerTopic() == nil {
		log.Println("MQTT consumer config not set.")
	}

	log.Printf("Connecting to MQTT broker at %s...", mqttClient.config.GetEndpoint())
	connection := MQTT.NewClient(opts)
	mqttClient.connection = connection
}

func (client *MQTTClient) Connect() error {
	if token := client.connection.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		return err
	}
	return nil
}

func (client *MQTTClient) Close() (err error) {
	log.Print("Closing MQTT connection...")
	client.connection.Unsubscribe(*client.config.GetConsumerTopic())
	client.connection.Disconnect(uint(10 / time.Millisecond))
	log.Print("MQTT connection closed.")
	return
}

func (client *MQTTClient) subscribe(handler func(message []byte)) {
	log.Printf("Subscribing to topic %s...", *client.config.GetConsumerTopic())
	h := func(client MQTT.Client, message MQTT.Message) {
		handler(message.Payload())
	}

	qos := client.config.GetConsumerQOS()
	if token := client.connection.Subscribe(*client.config.GetConsumerTopic(), byte(qos), h); token.Wait() && token.Error() != nil {
		log.Println(token)
	}
	log.Printf("Subscribed. Waiting for messages...")
}

func (client *MQTTClient) Subscribe(handler func(message []byte)) {
	client.handler = &handler
	client.subscribe(handler)
}

func (client *MQTTClient) Produce(message []byte) (err error) {
	topic := *client.config.GetProducerTopic()
	const qos = 1

	token := client.connection.Publish(topic, qos, true, message)
	token.Wait()
	err = token.Error()
	return
}
