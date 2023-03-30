package types

type MQTTConfig interface {
	ClientId() string
	Endpoint() string
	Username() string
	Password() string
	ProducerTopic() *string
	ConsumerTopic() *string
}
