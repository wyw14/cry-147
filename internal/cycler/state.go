package cycler

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type Channel struct {
	mu       sync.Mutex
	state    model.EquipmentState
	commands []model.ChannelCommand
}

func NewChannel(number int) *Channel {
	return &Channel{state: model.EquipmentState{Channel: number, Step: "idle", UpdatedAt: time.Now().UTC()}}
}

func (channel *Channel) Apply(command model.ChannelCommand) error {
	channel.mu.Lock()
	defer channel.mu.Unlock()
	if command.Channel != channel.state.Channel {
		return fmt.Errorf("command channel %d does not match %d", command.Channel, channel.state.Channel)
	}
	if channel.state.Protected && command.Kind != model.CommandStop {
		return fmt.Errorf("channel %d is protected", command.Channel)
	}
	switch command.Kind {
	case model.CommandSetCurrent:
		channel.state.SetCurrent = command.Value
		channel.state.Step = "current"
	case model.CommandSetVoltage:
		channel.state.SetVoltage = command.Value
		channel.state.Step = "voltage"
	case model.CommandPause:
		channel.state.Step = "paused"
	case model.CommandResume:
		channel.state.Step = "running"
	case model.CommandStop:
		channel.state.SetCurrent = 0
		channel.state.SetVoltage = 0
		channel.state.Step = "stopped"
		channel.state.Protected = true
	default:
		return fmt.Errorf("unsupported command %s", command.Kind)
	}
	channel.state.UpdatedAt = time.Now().UTC()
	channel.commands = append(channel.commands, command)
	return nil
}

func (channel *Channel) State() model.EquipmentState {
	channel.mu.Lock()
	defer channel.mu.Unlock()
	return channel.state
}

func (channel *Channel) Commands() []model.ChannelCommand {
	channel.mu.Lock()
	defer channel.mu.Unlock()
	return append([]model.ChannelCommand(nil), channel.commands...)
}

func (channel *Channel) RecordSequence(sequence uint64) {
	channel.mu.Lock()
	if sequence > channel.state.LastSequence {
		channel.state.LastSequence = sequence
		channel.state.UpdatedAt = time.Now().UTC()
	}
	channel.mu.Unlock()
}
