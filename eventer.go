package pubsub

type Eventer interface {
	GetTopic() string
	GetTopicID() string
}
