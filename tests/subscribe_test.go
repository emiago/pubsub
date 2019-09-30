package pubsub_test

import (
	"fmt"
	"pubsub"
	"pubsub/example"
	"sync"
	"testing"
)

func SubReceiveMessage(t *testing.T, sub *example.Subscriber, total int) int {
	count := 0
	for count < total {
		_, more := <-sub.Recv
		if !more {
			break
		}
		count++
	}

	return count
}

func TestUpdateSubscriberByEvent(t *testing.T) {
	subpool := pubsub.NewPool()
	subpool.Fulldebug = true
	go subpool.Run()
	var wg sync.WaitGroup
	total := 10

	id := fmt.Sprintf("PEER%d", 0)
	sub := example.NewSubscriber(id)
	sub.AddTopic("AGENTS", "1000")
	sub.AddTopic("AGENTS", "1001")
	subpool.AddSubscriber(sub)

	wg.Add(1)
	go func(sub *example.Subscriber) {
		defer wg.Done()
		count := SubReceiveMessage(t, sub, total)
		if count != total {
			t.Errorf("%s Did not receive all messages send=%d received=%d", sub.GetId(), total, count)
		}
	}(sub)

	//Lets fire some not subscribed events
	Nuns := 3
	for i := 0; i < Nuns; i++ {
		//Lets just use pubsub default events
		e := pubsub.NewEvent("AgentInfo", "AGENTS", "9999", "")
		subpool.QueueIt(&e)
	}

	subpool.QueueIt(&pubsub.SubUpdateEvent{
		Event: pubsub.Event{},
		SubId: id,
		Callback: func(subscriber pubsub.ISubscriber) {
			if subscriber == nil {
				return
			}

			//Because we are queueing messages, no external updating, this should work, otherwise we need to lock our pool
			// subpool.Lock()
			// defer subpool.Unlock()
			sub := subscriber.(*example.Subscriber)
			sub.RemoveTopic("AGENTS", "1000")
			sub.RemoveTopic("AGENTS", "1001")
			sub.AddTopic("AGENTS", "1002")
		},
	})

	for i := 0; i < total+10; i++ {
		//Lets just use pubsub default events
		e := pubsub.NewEvent("AgentInfo", "AGENTS", "1002", "")
		subpool.QueueIt(&e)
	}

	wg.Wait()

	//Currently subscriber should
	subscriber, exists := subpool.GetSubscriber(id)
	if !exists {
		t.Fatal("Subscriber does not exists", id)
	}

	sub = subscriber.(*example.Subscriber)
	if len(sub.Topics) > 1 {
		t.Fatal("Subscriber should have only 1 subscribtion")
	}

	if !sub.IsSubscribed("AGENTS", "1002") {
		t.Fatalf("Subscriber has wrong subscribtion %v, expected=%s", sub.Topics, "1002")
	}
}
