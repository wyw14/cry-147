package cycler

import (
	"context"

	"github.com/wyw14/cry-147/internal/model"
)

type DeviceGateway struct {
	coordinator *Coordinator
}

func NewDeviceGateway(coordinator *Coordinator) *DeviceGateway {
	return &DeviceGateway{coordinator: coordinator}
}

func (gateway *DeviceGateway) Connect(ctx context.Context, handshake Handshake) (*Lease, error) {
	return gateway.coordinator.Open(ctx, handshake)
}

func (gateway *DeviceGateway) SetCurrent(runID string, channel int, current float64) error {
	if _, err := gateway.coordinator.Schedule(runID, channel, model.CommandSetCurrent, current); err != nil {
		return err
	}
	_, _, err := gateway.coordinator.Dispatch(channel)
	return err
}

func (gateway *DeviceGateway) Equipment() []model.EquipmentState {
	return gateway.coordinator.States()
}

func (gateway *DeviceGateway) Recovery() Recovery {
	return gateway.coordinator.Recovery()
}
