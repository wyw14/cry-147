package profile

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type StageSink interface {
	Advance(string, model.Stage, string) (model.Operation, error)
}

type Integrator struct {
	runner    *Runner
	campaigns StageSink
}

func NewIntegrator(runner *Runner, campaigns StageSink) *Integrator {
	return &Integrator{runner: runner, campaigns: campaigns}
}

func (integrator *Integrator) Begin(runID string) error {
	if _, err := integrator.runner.BeginRest(); err != nil {
		return err
	}
	_, err := integrator.campaigns.Advance(runID, model.StageResting, "rest timer started")
	return err
}

func (integrator *Integrator) Pause() time.Duration {
	return integrator.runner.PauseRest()
}

func (integrator *Integrator) Resume() uint64 {
	return integrator.runner.ResumeRest()
}

func (integrator *Integrator) Checkpoint() Checkpoint {
	return integrator.runner.Checkpoint()
}

func (integrator *Integrator) Remaining() time.Duration {
	return integrator.runner.Remaining()
}

func (integrator *Integrator) Complete(ctx context.Context, runID string, temperature float64) (model.Operation, error) {
	if _, err := integrator.runner.Wait(ctx, temperature); err != nil {
		return model.Operation{}, fmt.Errorf("rest wait: %w", err)
	}
	return integrator.campaigns.Advance(runID, model.StageGrading, "rest completed")
}
