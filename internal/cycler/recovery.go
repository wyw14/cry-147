package cycler

import (
	"errors"
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type Recovery struct {
	States    []model.EquipmentState `json:"states"`
	Pending   map[int]int            `json:"pending"`
	Protected []int                  `json:"protected"`
}

func (coordinator *Coordinator) Recovery() Recovery {
	recovery := Recovery{States: coordinator.States(), Pending: map[int]int{}}
	coordinator.mu.RLock()
	defer coordinator.mu.RUnlock()
	for number, queue := range coordinator.queues {
		recovery.Pending[number] = queue.Pending()
		if queue.Protected() {
			recovery.Protected = append(recovery.Protected, number)
		}
	}
	return recovery
}

func (coordinator *Coordinator) Restore(states []model.EquipmentState) error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	for _, state := range states {
		channel, ok := coordinator.channels[state.Channel]
		if !ok {
			return fmt.Errorf("cannot restore channel %d", state.Channel)
		}
		channel.mu.Lock()
		channel.state = state
		channel.mu.Unlock()
		if state.Protected {
			coordinator.queues[state.Channel].protected = true
		}
	}
	return nil
}

func WrapAdapterError(channel int, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("cycler channel %d: %w", channel, err)
}

func Retryable(err error) bool {
	return errors.Is(err, ErrChannelBusy)
}
