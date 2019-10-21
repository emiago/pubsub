package pubsub_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/emiraganov/pubsub"
)

const (
	TOPIC1 = "TOPIC1"
	TOPIC2 = "TOPIC2"
	TOPIC3 = "TOPIC3"

	SUBSCRIBE_ALL_ID = "_ALL_"
)

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

//Lets implement ISubscriber interface
func (s *Subscriber) UID() string {
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
		TopicId: topicID,
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

func CheckSubReceivingMessage(t *testing.T, sub *Subscriber, total int) int {
	count := 0
	for count < total {
		m, more := <-sub.Recv
		if !more {
			break
		}
		//Test order of messages
		if m.GetTopicId() != fmt.Sprintf("%d", count) {
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
		if m.GetTopicId() != fmt.Sprintf("%d", count) {
			t.Fatalf("Order of messages not corrected")
		}
		count++
	}

	return count
}
