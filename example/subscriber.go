package example

import (
	"fmt"
	"pubsub"
)

const SUBSCRIBE_ALL_ID = "_ALL_"

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
func (s *Subscriber) GetId() string {
	return s.Id
}

func (s *Subscriber) IsSubscribed(top, topid string) bool {
	_, exists := s.Topics[FormatTopic(top, SUBSCRIBE_ALL_ID)]
	if !exists {
		_, exists = s.Topics[FormatTopic(top, topid)]
	}
	return exists
}

func (s *Subscriber) Send(m pubsub.Eventer) {
	select {
	case <-s.Closed:
	case s.Recv <- m:
	}
}

func (s *Subscriber) Close() {
	close(s.Closed)
}

func (s *Subscriber) AddTopic(top, topid string) {
	t := FormatTopic(top, topid)
	s.Topics[t]++
}

func FormatTopic(top, topid string) string {
	//Keep it simple
	return fmt.Sprintf("%s:%s", top, topid)
}
