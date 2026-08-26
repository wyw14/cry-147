package alarm

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type Recorder interface {
	Alarm(model.AlarmEvent) error
}

type Coordinator struct {
	evaluator *Evaluator
	state     *State
	recorder  Recorder
}

func NewCoordinator(evaluator *Evaluator, state *State, recorder Recorder) *Coordinator {
	return &Coordinator{evaluator: evaluator, state: state, recorder: recorder}
}

func (coordinator *Coordinator) Inspect(sample *model.Sample) (*model.AlarmEvent, error) {
	event, err := coordinator.evaluator.Evaluate(sample)
	if err != nil {
		return nil, fmt.Errorf("alarm evaluation failed: %w", err)
	}
	if event == nil {
		return nil, nil
	}
	coordinator.state.Apply(*event)
	if err := coordinator.recorder.Alarm(*event); err != nil {
		return nil, err
	}
	return event, nil
}

func (coordinator *Coordinator) Consume(ctx context.Context, stream <-chan model.AlarmEvent) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-stream:
			coordinator.state.Apply(event)
			if err := coordinator.recorder.Alarm(event); err != nil {
				return err
			}
		}
	}
}

func (coordinator *Coordinator) Active(runID string) []model.AlarmEvent {
	return coordinator.state.Active(runID)
}

func (coordinator *Coordinator) Recent(limit int) []model.AlarmEvent {
	return coordinator.state.History(limit)
}

func (coordinator *Coordinator) Latched(runID string) bool {
	return coordinator.state.Latched(runID)
}
