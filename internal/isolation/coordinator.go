package isolation

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type Coordinator struct {
	state      *State
	dispatcher *Dispatcher
}

func NewCoordinator(state *State, dispatcher *Dispatcher) *Coordinator {
	return &Coordinator{state: state, dispatcher: dispatcher}
}

func (coordinator *Coordinator) Latch(event model.AlarmEvent) error {
	if event.Cleared {
		return fmt.Errorf("cannot latch a cleared alarm")
	}
	coordinator.state.Latch(event)
	_, err := coordinator.dispatcher.Stop(event.RunID, event.Channel)
	return err
}

func (coordinator *Coordinator) Clear(event model.AlarmEvent) error {
	interlock, ok := coordinator.state.Clear(event)
	if !ok {
		return fmt.Errorf("interlock for run %s channel %d not found", event.RunID, event.Channel)
	}
	if interlock.Latched {
		return fmt.Errorf("interlock remained latched")
	}
	return nil
}

func (coordinator *Coordinator) List(runID string) []model.Interlock {
	return coordinator.state.List(runID)
}

func (coordinator *Coordinator) Protected(runID string, channel int) bool {
	for _, interlock := range coordinator.state.List(runID) {
		if interlock.Channel == channel && interlock.Latched {
			return true
		}
	}
	return false
}
