package pubsub_test

import (
	"fmt"
	"pubsub"
	"sync"
	"testing"
)

func SubReceiveMessage(t *testing.T, sub *Subscriber, total int) int {
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
	// log.SetLevel(log.DebugLevel)
	go subpool.Run()
	var wg sync.WaitGroup
	total := 10

	id := fmt.Sprintf("PEER%d", 0)
	sub := NewSubscriber(id)
	subpool.AddSubscriber(sub,
		"USERS:1000",
		"USERS:1001",
	)

	wg.Add(1)
	go func(sub *Subscriber) {
		defer wg.Done()
		count := SubReceiveMessage(t, sub, total)
		if count != total {
			t.Errorf("%s Did not receive all messages send=%d received=%d", sub.UID(), total, count)
		}
	}(sub)

	//Lets fire some not subscribed events
	for i := 0; i < 5; i++ {
		//Lets just use pubsub default events
		e := NewEvent("TopName", TOPIC1, "9999")
		subpool.QueueIt(&e)
	}

	subpool.QueueIt(&pubsub.SubUpdateEvent{
		Event: pubsub.Event{},
		SubId: id,
		Callback: func(subscriber pubsub.ISubscriber) {
			if subscriber == nil {
				return
			}

			// t.Log("Changing subscribtion")
			sub := subscriber.(*Subscriber)
			subpool.AddSubscriber(sub, "USERS:1002") //We need to add him again
			// log.Println(subpool.Stats())
		},
	})

	for i := 0; i < total+10; i++ {
		//Lets just use pubsub default events
		e := NewEvent("TopName", "USERS", "1002")
		subpool.QueueIt(&e)
	}

	// log.Println(subpool.Stats())

	wg.Wait()

	//Currently subscriber should

	if _, exists := subpool.GetSubscriber(id); !exists {
		t.Fatal("Subscriber does not exists", id)
	}

	if _, exists := subpool.GetSubscribersByTopic("USERS:1002"); !exists {
		t.Fatal("Subscriber has wrong subscribtion")
	}
}
