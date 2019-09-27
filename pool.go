package pubsub

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Pool struct {
	mu            sync.RWMutex
	subscribtions map[string]ISubscriber
	queue         chan Eventer
	log           logrus.FieldLogger
	Fulldebug     bool
}

func NewPool() *Pool {
	s := Pool{
		queue:         make(chan Eventer, 100), //
		subscribtions: make(map[string]ISubscriber),
		log:           logrus.WithFields(logrus.Fields{}),
		Fulldebug:     true,
	}
	return &s
}

func (r *Pool) SetLogger(l logrus.FieldLogger) {
	r.log = l
}

func (r *Pool) AddSubscriber(s ISubscriber) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := s.GetId()
	r.subscribtions[id] = s
	r.log.WithField("id", id).Info("Peer added in subpool")
}

func (r *Pool) RemoveSubscriber(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sub, exists := r.subscribtions[id]
	if exists {
		sub.Close() //Stop receiving any more messages
		delete(r.subscribtions, id)
	}
	r.log.WithField("id", id).Info("Peer removed from subpool")
}

func (r *Pool) Publish(e Eventer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, sub := range r.subscribtions {
		if !sub.IsSubscribed(e.GetTopic(), e.GetTopicId()) {
			continue
		}

		subid := sub.GetId()
		e.SetApplication(subid)

		if r.Fulldebug { //Following message can easily pill up logs, so it can be turned off
			r.log.WithFields(logrus.Fields{
				"sub":     subid,
				"event":   e.GetType(),
				"app":     e.GetApplication(),
				"topic":   e.GetTopic(),
				"topicid": e.GetTopicId(),
			}).Debug("Sub sending event ---->")
		}

		sub.Send(e)
	}
}

func (r *Pool) QueueIt(e Eventer) {
	r.queue <- e
}

func (r *Pool) Run() {
	r.log.Info("Starting subscription pool")
	for {
		e := <-r.queue
		switch m := e.(type) {
		case *SubRemoveEvent:
			r.RemoveSubscriber(m.SubId)
		default:
			r.Publish(e)
		}
	}
}
