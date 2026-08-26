package isolation

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/model"
)

type Dispatcher struct {
	devices *cycler.Coordinator
}

func NewDispatcher(devices *cycler.Coordinator) *Dispatcher {
	return &Dispatcher{devices: devices}
}

func (dispatcher *Dispatcher) Stop(runID string, channel int) (model.ChannelCommand, error) {
	command, err := dispatcher.devices.Schedule(runID, channel, model.CommandStop, 0)
	if err != nil {
		return model.ChannelCommand{}, err
	}
	dispatched, ok, err := dispatcher.devices.Dispatch(channel)
	if err != nil {
		return model.ChannelCommand{}, err
	}
	if !ok {
		return model.ChannelCommand{}, fmt.Errorf("no command dispatched for channel %d", channel)
	}
	if dispatched.ID != command.ID || dispatched.Kind != model.CommandStop {
		return model.ChannelCommand{}, fmt.Errorf("protective stop lost priority on channel %d", channel)
	}
	return dispatched, nil
}
