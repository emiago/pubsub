package pubsub_test

import (
	"fmt"
	"pubsub"
	"sync"
	"testing"
)

const (
	TOPIC1 = "TOPIC1"
	TOPIC2 = "TOPIC2"
	TOPIC3 = "TOPIC3"

	SUBSCRIBE_ALL_ID = "_ALL_"
)

func FormatTopic(top, topid string) string {
	//Keep it simple
	//Keep it simple
	if topid == SUBSCRIBE_ALL_ID {
		return top
	}

	return fmt.Sprintf("%s:%s", top, topid)
}

func NewEvent(name, topic, topicID string) pubsub.Event {
	return pubsub.Event{
		Topic:   topic,
		TopicID: topicID,
	}
}

func ConnSimulate(b *testing.B, ch chan pubsub.Eventer, stopCh chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	count := 0
	for {
		select {
		case <-ch:
			count++
		case <-stopCh:
			b.Log("Handled messages ", count)
			return
		}
	}
}

func CreateSubscriberToAll(id string) *Subscriber {
	sub := NewSubscriber(id)
	sub.AddTopic(TOPIC1, SUBSCRIBE_ALL_ID)
	sub.AddTopic(TOPIC2, SUBSCRIBE_ALL_ID)
	return sub
}

func CheckSubReceivingMessage(t *testing.T, sub *Subscriber, total int) int {
	count := 0
	for count < total {
		m, more := <-sub.Recv
		if !more {
			break
		}
		//Test order of messages
		if m.GetTopicID() != fmt.Sprintf("%d", count) {
			t.Fatalf("Order of messages not corrected")
		}
		count++
	}

	return count
}

func CheckSubReceivingMessageCount(t *testing.T, sub *Subscriber, total int) int {
	count := 0
	for count < total {
		m, more := <-sub.Recv
		if !more {
			break
		}
		//Test order of messages
		if m.GetTopicID() != fmt.Sprintf("%d", count) {
			t.Fatalf("Order of messages not corrected")
		}
		count++
	}

	return count
}

// This is our subscriber implementation used for testing

type Subscriber struct {
	Id     string //This is some id of subscriber which should be send as part of msg
	Topics map[string]int
	Recv   chan pubsub.Eventer
	Closed chan struct{}
}

func NewSubscriber(id string) *Subscriber {
	s := Subscriber{
		Id:     id,
		Recv:   make(chan pubsub.Eventer),
		Closed: make(chan struct{}),
		Topics: make(map[string]int),
	}

	return &s
}

func (s *Subscriber) Copy() *Subscriber {
	cp := NewSubscriber(s.Id)
	for k, v := range s.Topics {
		cp.Topics[k] = v
	}

	return cp
}

//Lets implement ISubscriber interface
func (s *Subscriber) GetId() string {
	return s.Id
}

func (s *Subscriber) GetTopics() []string {
	topics := make([]string, 0, len(s.Topics))
	for t, _ := range s.Topics {
		topics = append(topics, t)
	}

	return topics
}

func (s *Subscriber) Send(m pubsub.Eventer) error {
	select {
	case <-s.Closed:
	case s.Recv <- m:
	}

	return nil
}

func (s *Subscriber) Close() {
	close(s.Closed)
}

func (s *Subscriber) AddTopic(top, topid string) {
	t := FormatTopic(top, topid)
	s.Topics[t]++
}

func (s *Subscriber) RemoveTopic(top, topid string) {
	t := FormatTopic(top, topid)
	delete(s.Topics, t)
}

func (s *Subscriber) IsSubscribed(top, topid string) bool {
	t := FormatTopic(top, topid)
	_, exists := s.Topics[t]
	return exists
}
