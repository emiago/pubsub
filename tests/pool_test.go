package pubsub_test

import (
	"fmt"
	"io/ioutil"
	"pubsub"
	"pubsub/example"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

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

func CreateSubscriberToAll(id string) *example.Subscriber {
	sub := example.NewSubscriber(id)
	sub.AddTopic("AGENTS", example.SUBSCRIBE_ALL_ID)
	sub.AddTopic("CAMPAIGNS", example.SUBSCRIBE_ALL_ID)
	return sub
}

func CheckSubReceivingMessage(t *testing.T, sub *example.Subscriber, total int) int {
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

func BenchmarkPool(b *testing.B) {
	var wg sync.WaitGroup
	logrus.SetOutput(ioutil.Discard)
	// sendCh := make(chan interface{})
	stopConn := make(chan struct{})

	subpool := pubsub.NewPool()
	go subpool.Run()

	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("PEER%d", i)
		// fmt.Println("Subscribing peer", id)
		sub := CreateSubscriberToAll(id)
		subpool.AddSubscriber(sub)
		wg.Add(1)
		go ConnSimulate(b, sub.Recv, stopConn, &wg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//Lets just use pubsub default events
		e := pubsub.NewEvent("AgentInfo", "AGENTS", "123", "")
		// fmt.Println("queueign")
		subpool.QueueIt(&e)
	}
	close(stopConn)
	wg.Wait()
}

func TestPoolGettingMessage(t *testing.T) {
	subpool := pubsub.NewPool()
	subpool.Fulldebug = true
	go subpool.Run()

	id := fmt.Sprintf("TESTER")
	sub := CreateSubscriberToAll(id)
	subpool.AddSubscriber(sub)

	total := 10
	go func() {
		for i := 0; i < total; i++ {
			//Lets just use pubsub default events
			e := pubsub.NewEvent("AgentInfo", "AGENTS", fmt.Sprintf("%d", i), "")
			// fmt.Println("queueign")
			subpool.QueueIt(&e)
		}
	}()

	count := CheckSubReceivingMessage(t, sub, total)

	if count != total {
		t.Errorf("Did not receive all messages send=%d received=%d", total, count)
	}

	t.Logf("Messages sent=%d received=%d", total, count)
}

func TestPoolNSubscribersGettingMessage(t *testing.T) {
	subpool := pubsub.NewPool()
	subpool.Fulldebug = true
	go subpool.Run()
	var wg sync.WaitGroup
	total := 10

	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("PEER%d", i)
		// fmt.Println("Subscribing peer", id)
		sub := CreateSubscriberToAll(id)
		subpool.AddSubscriber(sub)

		wg.Add(1)
		go func(sub *example.Subscriber, wg *sync.WaitGroup) {
			defer wg.Done()
			count := CheckSubReceivingMessage(t, sub, total)
			if count != total {
				t.Errorf("%s Did not receive all messages send=%d received=%d", sub.GetId(), total, count)
			}
		}(sub, &wg)
	}

	for i := 0; i < total; i++ {
		//Lets just use pubsub default events
		e := pubsub.NewEvent("AgentInfo", "AGENTS", fmt.Sprintf("%d", i), "")
		subpool.QueueIt(&e)
	}

	wg.Wait()
}
