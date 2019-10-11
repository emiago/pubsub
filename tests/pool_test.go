package pubsub_test

import (
	"fmt"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/emiraganov/pubsub"

	"github.com/sirupsen/logrus"
)

func TestPoolSubUnsub(t *testing.T) {
	subpool := pubsub.NewPool()
	subpool.Fulldebug = true
	logrus.SetLevel(logrus.DebugLevel)
	go subpool.Run()

	id := fmt.Sprintf("TESTER")
	sub := NewSubscriber(id)
	subpool.AddSubscriber(sub,
		"TOPIC1:1000",
		"TOPIC1:1001",
	)

	if _, exists := subpool.GetSubscribersByTopic("TOPIC1:1000"); !exists {
		t.Error("Subscribtion failed")
	}

	t.Log(subpool.Stats())

	subpool.AddSubscriber(sub,
		"TOPIC1:1000",
	)

	if _, exists := subpool.GetSubscribersByTopic("TOPIC1:1001"); exists {
		t.Error("Unsubscribe failed")
	}

	t.Log(subpool.Stats())
}

func TestPoolGettingMessage(t *testing.T) {
	subpool := pubsub.NewPool()
	subpool.Fulldebug = true
	go subpool.Run()

	id := fmt.Sprintf("TESTER")
	sub := NewSubscriber(id)
	subpool.AddSubscriber(sub,
		"FIRE",
		"WATER",
		"EARTH",
	)

	total := 10
	go func() {
		for i := 0; i < total; i++ {
			//Lets just use pubsub default events
			e := NewEvent("PoolBurn", "FIRE", fmt.Sprintf("%d", i))
			// fmt.Println("queueign")
			subpool.QueueIt(&e)
		}
	}()

	count := CheckSubReceivingMessageCount(t, sub, total)

	if count != total {
		t.Errorf("Did not receive all messages send=%d received=%d", total, count)
	}

	t.Logf("Messages sent=%d received=%d", total, count)
	t.Log(subpool.Stats())
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
		sub := NewSubscriber(id)
		subpool.AddSubscriber(sub, "FIRE", "WATER")

		wg.Add(1)
		go func(sub *Subscriber, wg *sync.WaitGroup) {
			defer wg.Done()
			count := CheckSubReceivingMessage(t, sub, total)
			if count != total {
				t.Errorf("%s Did not receive all messages send=%d received=%d", sub.UID(), total, count)
			}
		}(sub, &wg)
	}

	for i := 0; i < total; i++ {
		//Lets just use pubsub default events
		e := NewEvent("PoolFire", "FIRE", fmt.Sprintf("%d", i))
		subpool.QueueIt(&e)
	}

	wg.Wait()
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
		sub := NewSubscriber(id)
		subpool.AddSubscriber(sub, "FIRE", "WATER")
		wg.Add(1)
		go ConnSimulate(b, sub.Recv, stopConn, &wg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//Lets just use pubsub default events
		e := NewEvent("PoolFire", "FIRE", "123")
		// fmt.Println("queueign")
		subpool.QueueIt(&e)
	}
	close(stopConn)
	wg.Wait()
}
