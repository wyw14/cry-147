package cycler

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/wyw14/cry-147/internal/model"
)

var ErrChannelBusy = errors.New("channel busy")

type Coordinator struct {
	mu        sync.RWMutex
	channels  map[int]*Channel
	queues    map[int]*CommandQueue
	admission *Admission
}

func NewCoordinator(channelCount int, sessionLimit int) *Coordinator {
	coordinator := &Coordinator{channels: map[int]*Channel{}, queues: map[int]*CommandQueue{}, admission: NewAdmission(sessionLimit)}
	for number := 1; number <= channelCount; number++ {
		coordinator.channels[number] = NewChannel(number)
		coordinator.queues[number] = NewCommandQueue()
	}
	return coordinator
}

func (coordinator *Coordinator) Open(ctx context.Context, handshake Handshake) (*Lease, error) {
	lease, err := coordinator.admission.Open(ctx, handshake)
	if err != nil {
		return nil, fmt.Errorf("open channel session: %w", err)
	}
	return lease, nil
}

func (coordinator *Coordinator) Schedule(runID string, channel int, kind model.CommandKind, value float64) (model.ChannelCommand, error) {
	coordinator.mu.RLock()
	queue, ok := coordinator.queues[channel]
	coordinator.mu.RUnlock()
	if !ok {
		return model.ChannelCommand{}, fmt.Errorf("channel %d not found", channel)
	}
	return queue.Enqueue(runID, channel, kind, value)
}

func (coordinator *Coordinator) Dispatch(channel int) (model.ChannelCommand, bool, error) {
	coordinator.mu.RLock()
	queue, queueOK := coordinator.queues[channel]
	device, deviceOK := coordinator.channels[channel]
	coordinator.mu.RUnlock()
	if !queueOK || !deviceOK {
		return model.ChannelCommand{}, false, fmt.Errorf("channel %d not found", channel)
	}
	command, ok := queue.Next()
	if !ok {
		return model.ChannelCommand{}, false, nil
	}
	if err := device.Apply(command); err != nil {
		return command, true, err
	}
	return command, true, nil
}

func (coordinator *Coordinator) States() []model.EquipmentState {
	coordinator.mu.RLock()
	numbers := make([]int, 0, len(coordinator.channels))
	for number := range coordinator.channels {
		numbers = append(numbers, number)
	}
	coordinator.mu.RUnlock()
	sort.Ints(numbers)
	out := make([]model.EquipmentState, 0, len(numbers))
	for _, number := range numbers {
		coordinator.mu.RLock()
		channel := coordinator.channels[number]
		coordinator.mu.RUnlock()
		out = append(out, channel.State())
	}
	return out
}

func (coordinator *Coordinator) AdmissionUsage() (int, int) {
	return coordinator.admission.Used(), coordinator.admission.Active()
}

func (coordinator *Coordinator) RecordSequence(channel int, sequence uint64) error {
	coordinator.mu.RLock()
	device, ok := coordinator.channels[channel]
	coordinator.mu.RUnlock()
	if !ok {
		return fmt.Errorf("channel %d not found", channel)
	}
	device.RecordSequence(sequence)
	return nil
}

func (coordinator *Coordinator) Commands(channel int) ([]model.ChannelCommand, error) {
	coordinator.mu.RLock()
	device, ok := coordinator.channels[channel]
	coordinator.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("channel %d not found", channel)
	}
	return device.Commands(), nil
}
