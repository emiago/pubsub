package pubsub

type ISubscriber interface {
	GetId() string
	IsSubscribed(topic, topicid string) bool
	Send(e Eventer) //This is called if peer is subscribed
}
