package sample

import (
	"fmt"
	"sort"
	"sync"

	"github.com/wyw14/cry-147/internal/model"
)

type Subscriber func(*model.Sample) error

type Coordinator struct {
	mu          sync.RWMutex
	subscribers map[string]Subscriber
	pool        sync.Pool
}

func NewCoordinator() *Coordinator {
	coordinator := &Coordinator{subscribers: map[string]Subscriber{}}
	coordinator.pool.New = func() any { return &model.Sample{} }
	return coordinator
}

func (coordinator *Coordinator) Subscribe(id string, subscriber Subscriber) func() {
	coordinator.mu.Lock()
	coordinator.subscribers[id] = subscriber
	coordinator.mu.Unlock()
	return func() {
		coordinator.mu.Lock()
		delete(coordinator.subscribers, id)
		coordinator.mu.Unlock()
	}
}

func (coordinator *Coordinator) subscriberList() []Subscriber {
	coordinator.mu.RLock()
	ids := make([]string, 0, len(coordinator.subscribers))
	for id := range coordinator.subscribers {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Subscriber, 0, len(ids))
	for _, id := range ids {
		out = append(out, coordinator.subscribers[id])
	}
	coordinator.mu.RUnlock()
	return out
}

func (coordinator *Coordinator) Publish(sample *model.Sample) error {
	if sample == nil {
		return fmt.Errorf("sample is required")
	}
	subscribers := coordinator.subscriberList()
	errorsByIndex := make([]error, len(subscribers))
	var group sync.WaitGroup
	for index, subscriber := range subscribers {
		copySample := sample.Clone()
		group.Add(1)
		go func(position int, handler Subscriber, value *model.Sample) {
			defer group.Done()
			errorsByIndex[position] = handler(value)
		}(index, subscriber, copySample)
	}
	group.Wait()
	for _, err := range errorsByIndex {
		if err != nil {
			return err
		}
	}
	return nil
}

func (coordinator *Coordinator) PublishPooled(input model.Sample) error {
	pooled := coordinator.pool.Get().(*model.Sample)
	*pooled = input
	pooled.Payload = model.CloneBytes(input.Payload)
	err := coordinator.Publish(pooled)
	*pooled = model.Sample{}
	coordinator.pool.Put(pooled)
	return err
}

func (coordinator *Coordinator) SubscriberCount() int {
	coordinator.mu.RLock()
	defer coordinator.mu.RUnlock()
	return len(coordinator.subscribers)
}
