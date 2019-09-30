package pubsub

import "time"

type Eventer interface {
	GetType() string
	GetTopic() string
	GetTopicId() string
	GetApplication() string
	SetApplication(t string)
}

type Event struct {
	Type        string `json:"type"`
	Application string `json:"application"`
	Topic       string `json:"topic"`
	TopicId     string `json:"topic_id"`
	Timestamp   *Time  `json:"timestamp"`
}

func NewEvent(t string, topic string, topicid string, app string) Event {
	time := Time(time.Now())

	return Event{
		Type:        t,
		Application: app,
		Topic:       topic,
		TopicId:     topicid,
		Timestamp:   &time,
	}
}

func (p *Event) GetType() string {
	return p.Type
}

func (p *Event) GetTopic() string {
	return p.Topic
}

func (p *Event) GetTopicId() string {
	return p.TopicId
}

func (p *Event) GetApplication() string {
	return p.Application
}

func (p *Event) SetApplication(t string) {
	p.Application = t
}
