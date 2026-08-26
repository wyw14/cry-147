package quality

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/tray"
)

type Journal interface {
	Sample(*model.Sample) error
	Grade(string, model.GradeResult) error
}

type Integrator struct {
	quality *Coordinator
	trays   *tray.SampleIntegrator
	journal Journal
}

func NewIntegrator(quality *Coordinator, trays *tray.SampleIntegrator, journal Journal) *Integrator {
	return &Integrator{quality: quality, trays: trays, journal: journal}
}

func (integrator *Integrator) Sample(sample *model.Sample) error {
	if sample == nil {
		return fmt.Errorf("sample is required")
	}
	if err := integrator.quality.Accept(sample); err != nil {
		return err
	}
	if err := integrator.trays.Apply(sample); err != nil {
		return err
	}
	return integrator.journal.Sample(sample)
}

func (integrator *Integrator) Publish(runID string, trayID string) (model.GradeResult, error) {
	result, err := integrator.quality.Publish(runID, trayID)
	if err != nil {
		return model.GradeResult{}, err
	}
	if err := integrator.journal.Grade(runID, result); err != nil {
		return model.GradeResult{}, err
	}
	return result.Clone(), nil
}

func (integrator *Integrator) Preview(runID string, trayID string) (model.GradeResult, *model.Tray, error) {
	return integrator.quality.Preview(runID, trayID)
}
