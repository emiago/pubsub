package pubsub

type ISubscriber interface {
	GetId() string
	Send(e Eventer) error //This is called if peer is subscribed
}

type Subscriber struct {
	sub    ISubscriber
	topics []string
}
