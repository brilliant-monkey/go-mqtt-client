package types

type MQTTConfig interface {
	GetClientId() string
	GetEndpoint() string
	GetUsername() string
	GetPassword() string
	GetProducerTopic() *string
	GetConsumerTopic() *string
}
