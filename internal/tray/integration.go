package tray

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type SampleIntegrator struct {
	registry *Registry
}

func NewSampleIntegrator(registry *Registry) *SampleIntegrator {
	return &SampleIntegrator{registry: registry}
}

func (integrator *SampleIntegrator) Apply(sample *model.Sample) error {
	if sample == nil {
		return fmt.Errorf("sample is required")
	}
	tray, ok := integrator.registry.Lookup(sample.TrayID)
	if !ok {
		return fmt.Errorf("tray %s not found", sample.TrayID)
	}
	if err := UpdateCellSample(tray, sample); err != nil {
		return err
	}
	return integrator.registry.Replace(tray)
}

func (integrator *SampleIntegrator) Curve(trayID string) ([]model.Cell, error) {
	tray, ok := integrator.registry.Lookup(trayID)
	if !ok {
		return nil, fmt.Errorf("tray %s not found", trayID)
	}
	out := make([]model.Cell, 0, len(tray.Cells))
	for _, id := range model.SortedCellIDs(tray) {
		out = append(out, *tray.Cells[id].Clone())
	}
	return out, nil
}

func (integrator *SampleIntegrator) Summary(runID string) model.RecoveryReport {
	return model.BuildRecoveryReport(runID, integrator.registry.List())
}
