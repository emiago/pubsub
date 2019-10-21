package pubsub

type Eventer interface {
	GetTopic() string
	GetTopicId() string
}
