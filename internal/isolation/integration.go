package isolation

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/tray"
)

type Integrator struct {
	isolation *Coordinator
	trays     *tray.Registry
}

func NewIntegrator(isolation *Coordinator, trays *tray.Registry) *Integrator {
	return &Integrator{isolation: isolation, trays: trays}
}

func (integrator *Integrator) Latch(event model.AlarmEvent) error {
	live, ok := integrator.trays.Lookup(event.TrayID)
	if !ok {
		return fmt.Errorf("tray %s not found", event.TrayID)
	}
	if err := tray.IsolateCell(live, event.CellID); err != nil {
		return err
	}
	if err := integrator.trays.Replace(live); err != nil {
		return err
	}
	return integrator.isolation.Latch(event)
}

func (integrator *Integrator) Clear(event model.AlarmEvent) error {
	live, ok := integrator.trays.Lookup(event.TrayID)
	if !ok {
		return fmt.Errorf("tray %s not found", event.TrayID)
	}
	if err := tray.ClearIsolation(live, event.CellID); err != nil {
		return err
	}
	if err := integrator.trays.Replace(live); err != nil {
		return err
	}
	return integrator.isolation.Clear(event)
}

func (integrator *Integrator) Interlocks(runID string) []model.Interlock {
	return integrator.isolation.List(runID)
}
