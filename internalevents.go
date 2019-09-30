package pubsub

type SubUpdateEvent struct {
	Event
	SubId    string
	Callback func(sub ISubscriber)
}
