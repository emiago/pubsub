package main

/*
	Example of some contact center server, where you can listen
	on Agents and Queues
*/

import (
	"context"
	"fmt"
	"time"

	// "log"
	"net"
	"github.com/emiraganov/pubsub"

	pb "github.com/emiraganov/pubsub/example/grpcserver/mymodels"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

var (
	subpool *pubsub.Pool
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

func (s *server) GetEvents(req *pb.Subscribe, srv pb.Greeter_GetEventsServer) error {
	log.Println("User subscribed", req.App)

	sub := NewSubscriber(req.App)
	defer sub.Close()

	// var topics []string
	for _, t := range req.Agents {
		sub.AddTopic("AGENTS", t)
	}

	for _, t := range req.Queues {
		sub.AddTopic("QUEUES", t)
	}

	subpool.AddSubscriber(sub, sub.GetTopics()...)

	ctx := srv.Context()

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for i := 0; i < 10; i++ {
			// data.Value = []byte("{some json will be here}")
			event := &pb.TestEvent{Somestuff: "This is just example data"}
			queueEvent("TestEvent", "AGENTS", "AGENTS1000", sub.Id, event)
			time.Sleep(1 * time.Second)
		}
	}(ctx)

	log.Println(subpool.Stats())

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelation")
			return nil
		case m := <-sub.Recv:
			if err := srv.SendMsg(m); err != nil {
				return err
			}
		}
	}

	// return nil
}

func queueEvent(name, topic, topicID, app string, data proto.Message) {
	serialized, _ := ptypes.MarshalAny(data)
	subpool.QueueIt(&pb.Event{
		Type:        name,
		Topic:       topic,
		TopicID:     topicID,
		Application: app,
		Data:        serialized,
	})
}

func main() {
	subpool = pubsub.NewPool()
	subpool.Fulldebug = true
	// log.SetLevel(log.DebugLevel)
	subpool.SetLogger(log.WithFields(log.Fields{"cmd": "POOL"}))
	go subpool.Run()

	log.Debug("TESTGING")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

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
	//We are using channels to avoid blocking pool messenger
	select {
	case <-s.Closed:
		return fmt.Errorf("I am closed, you can log this")
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

func FormatTopic(top, topid string) string {
	//Keep it simple
	if topid == SUBSCRIBE_ALL_ID {
		return top
	}

	return fmt.Sprintf("%s:%s", top, topid)
}
