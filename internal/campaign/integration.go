package campaign

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/cycler"
)

type EquipmentIntegrator struct {
	campaigns *Coordinator
	devices   *cycler.Coordinator
}

func NewEquipmentIntegrator(campaigns *Coordinator, devices *cycler.Coordinator) *EquipmentIntegrator {
	return &EquipmentIntegrator{campaigns: campaigns, devices: devices}
}

func (integrator *EquipmentIntegrator) ApplyResult(runID string, err error) (bool, error) {
	integrator.campaigns.mu.RLock()
	state, ok := integrator.campaigns.states[runID]
	integrator.campaigns.mu.RUnlock()
	if !ok {
		return false, fmt.Errorf("run %s not found", runID)
	}
	wrapped := cycler.WrapAdapterError(0, err)
	retry := state.DeviceResult(wrapped, cycler.ErrChannelBusy)
	if retry != cycler.Retryable(wrapped) {
		return false, fmt.Errorf("retry classification disagrees with campaign state")
	}
	return retry, nil
}

func (integrator *EquipmentIntegrator) RetryCount(runID string) int {
	integrator.campaigns.mu.RLock()
	state := integrator.campaigns.states[runID]
	integrator.campaigns.mu.RUnlock()
	if state == nil {
		return 0
	}
	return state.RetryCount()
}
