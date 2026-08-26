package cycler

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

type commandItem struct {
	command model.ChannelCommand
	order   uint64
}

type commandHeap []commandItem

func (items commandHeap) Len() int { return len(items) }
func (items commandHeap) Less(i int, j int) bool {
	// Higher priority commands (e.g. protective stops) must overtake the
	// backlog of routine current/voltage tweaks that were enqueued before the
	// isolation latch. Compare by priority first, then fall back to FIFO order
	// so commands of equal priority are dispatched in the order they arrived.
	if items[i].command.Priority != items[j].command.Priority {
		return items[i].command.Priority > items[j].command.Priority
	}
	return items[i].order < items[j].order
}
func (items commandHeap) Swap(i int, j int) { items[i], items[j] = items[j], items[i] }
func (items *commandHeap) Push(value any)   { *items = append(*items, value.(commandItem)) }
func (items *commandHeap) Pop() any {
	old := *items
	last := old[len(old)-1]
	*items = old[:len(old)-1]
	return last
}

type CommandQueue struct {
	mu        sync.Mutex
	items     commandHeap
	next      uint64
	protected bool
}

func NewCommandQueue() *CommandQueue {
	queue := &CommandQueue{}
	heap.Init(&queue.items)
	return queue
}

func (queue *CommandQueue) Enqueue(runID string, channel int, kind model.CommandKind, value float64) (model.ChannelCommand, error) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.protected && kind != model.CommandStop {
		return model.ChannelCommand{}, fmt.Errorf("channel %d rejects %s while protected", channel, kind)
	}
	priority := 10
	if kind == model.CommandStop {
		priority = 1000
		queue.protected = true
	}
	command := model.ChannelCommand{ID: uuid.NewString(), RunID: runID, Channel: channel, Kind: kind, Value: value, Priority: priority, CreatedAt: time.Now().UTC()}
	queue.next++
	heap.Push(&queue.items, commandItem{command: command, order: queue.next})
	return command, nil
}

func (queue *CommandQueue) Next() (model.ChannelCommand, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.items) == 0 {
		return model.ChannelCommand{}, false
	}
	item := heap.Pop(&queue.items).(commandItem)
	return item.command, true
}

func (queue *CommandQueue) Pending() int {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return len(queue.items)
}

func (queue *CommandQueue) Protected() bool {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return queue.protected
}
