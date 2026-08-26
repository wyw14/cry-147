package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/profile"
	"github.com/wyw14/cry-147/internal/thermal"
	"github.com/wyw14/cry-147/internal/tray"
)

type RestStatus struct {
	RunID      string             `json:"run_id"`
	Checkpoint profile.Checkpoint `json:"checkpoint"`
	Remaining  time.Duration      `json:"remaining"`
}

func (system *System) PauseRest(runID string) (RestStatus, error) {
	system.mu.RLock()
	integrator := system.profiles[runID]
	system.mu.RUnlock()
	if integrator == nil {
		return RestStatus{}, fmt.Errorf("rest profile for run %s not found", runID)
	}
	remaining := integrator.Pause()
	return RestStatus{RunID: runID, Checkpoint: integrator.Checkpoint(), Remaining: remaining}, nil
}

func (system *System) ResumeRest(runID string) (RestStatus, error) {
	system.mu.RLock()
	integrator := system.profiles[runID]
	system.mu.RUnlock()
	if integrator == nil {
		return RestStatus{}, fmt.Errorf("rest profile for run %s not found", runID)
	}
	integrator.Resume()
	return RestStatus{RunID: runID, Checkpoint: integrator.Checkpoint(), Remaining: integrator.Remaining()}, nil
}

func (system *System) RestStatus(runID string) (RestStatus, error) {
	system.mu.RLock()
	integrator := system.profiles[runID]
	system.mu.RUnlock()
	if integrator == nil {
		return RestStatus{}, fmt.Errorf("rest profile for run %s not found", runID)
	}
	return RestStatus{RunID: runID, Checkpoint: integrator.Checkpoint(), Remaining: integrator.Remaining()}, nil
}

func (system *System) ClearIsolation(runID string, trayID string, cellID string, channel int, code string) error {
	if !system.isolation.Protected(runID, channel) {
		return fmt.Errorf("interlock for run %s channel %d is not latched", runID, channel)
	}
	return system.alarmFlow.Clear(runID, trayID, cellID, channel, code)
}

func (system *System) EquipmentResult(runID string, channel int, busy bool, message string) (map[string]any, error) {
	var result error
	if busy {
		result = cycler.ErrChannelBusy
	} else if message != "" {
		result = errors.New(message)
	}
	retry, err := system.equipment.ApplyResult(runID, result)
	if err != nil {
		return nil, err
	}
	commands, err := system.devices.Commands(channel)
	if err != nil {
		return nil, err
	}
	return map[string]any{"retry": retry, "retry_count": system.equipment.RetryCount(runID), "commands": commands}, nil
}

func (system *System) GradeResult(runID string) (model.GradeResult, bool) {
	return system.quality.Result(runID)
}

func (system *System) ThermalReading(zone string) (thermal.Reading, bool) {
	return system.thermalState.Reading(zone)
}

func (system *System) WaitThermalStable(ctx context.Context, zone string, interval time.Duration) error {
	return system.thermalFlow.WaitStable(ctx, zone, interval)
}

func (system *System) IsolatedChannels(trayID string) ([]int, error) {
	value, ok := system.trays.Lookup(trayID)
	if !ok {
		return nil, fmt.Errorf("tray %s not found", trayID)
	}
	return tray.IsolatedChannels(value), nil
}
