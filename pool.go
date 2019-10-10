package pubsub

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
)

type Pool struct {
	mu            sync.RWMutex
	subscribtions map[string]*Subscriber
	topics        map[string][]*Subscriber
	queue         chan Eventer
	log           logrus.FieldLogger
	Fulldebug     bool
}

func NewPool() *Pool {
	s := Pool{
		queue:         make(chan Eventer, 100), //
		subscribtions: make(map[string]*Subscriber),
		topics:        make(map[string][]*Subscriber),
		log:           logrus.WithFields(logrus.Fields{}),
		Fulldebug:     true,
	}
	return &s
}

//For external use
func (r *Pool) Lock() {
	r.mu.Lock()
}

//For external use
func (r *Pool) Unlock() {
	r.mu.Unlock()
}

func (r *Pool) SetLogger(l logrus.FieldLogger) {
	r.log = l
}

func (r *Pool) addSub(sub ISubscriber, topics []string) {
	s := &Subscriber{
		sub: sub,
	}
	s.topics = append(s.topics, topics...) //Copy topics

	r.subscribtions[sub.UID()] = s
	for _, t := range s.topics {
		r.topics[t] = append(r.topics[t], s)
	}
}

func (r *Pool) removeSub(id string) *Subscriber {
	s, exists := r.subscribtions[id]
	if !exists {
		return nil
	}

	r.unsubscribe(s, s.topics)
	delete(r.subscribtions, id)
	return s
}

func (r *Pool) unsubscribe(s *Subscriber, topics []string) {
	for _, t := range topics {
		subs := r.topics[t]
		for i, sub := range subs {
			if s == sub {
				subs = append(subs[:i], subs[i+1:]...)
				break
			}
		}

		if len(subs) == 0 {
			delete(r.topics, t)
			continue
		}

		r.topics[t] = subs
	}
}

// AddSubscriber - adds subscriber into pool, it can act as update also
func (r *Pool) AddSubscriber(s ISubscriber, topics ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := s.UID()
	r.removeSub(id) //Remove existing subscribtion
	r.addSub(s, topics)
	r.log.WithField("id", id).Info("Subscriber added in subpool")
}

func (r *Pool) RemoveSubscriber(id string) ISubscriber {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.removeSub(id)
	r.log.WithField("id", id).Info("Subscriber removed from subpool")
	return s.sub
}

func (r *Pool) GetSubscriber(id string) (ISubscriber, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, exists := r.subscribtions[id]
	return s.sub, exists
}

func (r *Pool) GetSubscribersByTopic(topic string) ([]ISubscriber, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	subs, exists := r.topics[topic]
	var list []ISubscriber

	for _, s := range subs {
		list = append(list, s.sub)
	}

	return list, exists
}

func (r *Pool) QueueIt(e Eventer) {
	r.queue <- e
}

func (r *Pool) Run() {
	r.log.Info("Starting subscription pool")
	for {
		e := <-r.queue
		switch m := e.(type) {
		case *SubUpdateEvent:
			s, _ := r.GetSubscriber(m.SubId)
			m.Callback(s)
		default:
			r.Publish(e)
		}
	}
}

func (r *Pool) Publish(e Eventer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Fulldebug {
		r.log.WithFields(logrus.Fields{
			"topic":   e.GetTopic(),
			"topicID": e.GetTopicID(),
		}).Debug("Publishing event ---->")
	}

	//Brodcasting to all subscribers for current topic. Topic can be generic also
	topic := e.GetTopic()
	for _, s := range r.topics[topic] {
		r.subSend(s.sub, e)
	}

	//Brodcasting to all subscribers for current topic ID
	topicID := fmt.Sprintf("%s:%s", e.GetTopic(), e.GetTopicID())
	for _, s := range r.topics[topicID] {
		r.subSend(s.sub, e)
	}
}

func (r *Pool) subSend(sub ISubscriber, e Eventer) {
	if r.Fulldebug {
		r.log.WithFields(logrus.Fields{
			"sub":     sub.UID(),
			"topic":   e.GetTopic(),
			"topicID": e.GetTopicID(),
		}).Debug("Sub sending event ---->")
	}

	//This is not safe sending, as event will be pointer. SUBSCRIBERS SHOULD not change message,
	//if that is required then make sure your susbcriber encodes or copies event
	if err := sub.Send(e); err != nil {
		//Here we could do special error handling like
		//If this error is that then maybe resend this message
		r.log.WithError(err).Debug("Message was not sent")
	}
}

func (r *Pool) Stats() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := `Pool stats

Subscribers
%v

Topics
%v

Queue
	size: %d,
	messages: %d,
`
	return fmt.Sprintf(out,
		r.subscribtions,
		r.topics,
		cap(r.queue),
		len(r.queue),
	)
}
