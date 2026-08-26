package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/profile"
	"github.com/wyw14/cry-147/internal/thermal"
)

func (system *System) Isolate(runID string, trayID string, cellID string, channel int, reason string) error {
	event := model.AlarmEvent{ID: uuid.NewString(), RunID: runID, TrayID: trayID, CellID: cellID, Channel: channel, Level: model.AlarmCritical, Code: "manual-interlock", Message: reason, OccurredAt: time.Now().UTC()}
	if err := system.isolationFlow.Latch(event); err != nil {
		return err
	}
	system.RaiseIncident(runID, trayID, reason, model.AlarmCritical)
	return nil
}

func (system *System) PreviewGrade(runID string, trayID string) (model.GradeResult, *model.Tray, error) {
	return system.qualityFlow.Preview(runID, trayID)
}

func (system *System) PublishGrade(runID string, trayID string) (model.GradeResult, error) {
	return system.qualityFlow.Publish(runID, trayID)
}

func (system *System) StartRest(runID string, duration time.Duration, minimum float64, maximum float64) error {
	state, err := profile.NewState([]profile.Step{{Name: "rest", Duration: duration, MinimumTemperature: minimum, MaximumTemperature: maximum}, {Name: "grade", Duration: time.Millisecond, MinimumTemperature: minimum, MaximumTemperature: maximum}})
	if err != nil {
		return err
	}
	integrator := profile.NewIntegrator(profile.NewRunner(state, profile.NewRestTimer()), system.campaigns)
	if err := integrator.Begin(runID); err != nil {
		return err
	}
	system.mu.Lock()
	system.profiles[runID] = integrator
	system.mu.Unlock()
	return nil
}

func (system *System) CompleteRest(ctx context.Context, runID string, temperature float64) (model.Operation, error) {
	system.mu.RLock()
	integrator := system.profiles[runID]
	system.mu.RUnlock()
	if integrator == nil {
		return model.Operation{}, fmt.Errorf("rest profile for run %s not found", runID)
	}
	return integrator.Complete(ctx, runID, temperature)
}

func (system *System) PollThermal(ctx context.Context, zone string) (thermal.Reading, error) {
	return system.thermalFlow.Poll(ctx, zone)
}
