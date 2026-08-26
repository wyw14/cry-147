package quality

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/tray"
)

type Coordinator struct {
	state  *State
	trays  *tray.Registry
	grader *Grader
}

func NewCoordinator(state *State, trays *tray.Registry, grader *Grader) *Coordinator {
	return &Coordinator{state: state, trays: trays, grader: grader}
}

func (coordinator *Coordinator) Accept(sample *model.Sample) error {
	return coordinator.state.Accept(sample)
}

func (coordinator *Coordinator) Preview(runID string, trayID string) (model.GradeResult, *model.Tray, error) {
	live, ok := coordinator.trays.Lookup(trayID)
	if !ok {
		return model.GradeResult{}, nil, fmt.Errorf("tray %s not found", trayID)
	}
	snapshot, err := tray.Freeze(live)
	if err != nil {
		return model.GradeResult{}, nil, err
	}
	previewSnapshot := snapshot.Clone()
	if len(previewSnapshot.CellOrder) > 0 {
		if _, err := previewSnapshot.Cell(previewSnapshot.CellOrder[0]); err != nil {
			return model.GradeResult{}, nil, err
		}
	}
	_ = previewSnapshot.Capacities()
	preview := previewSnapshot.Tray
	result, err := coordinator.grader.Calculate(preview, coordinator.state.Samples(runID))
	if err != nil {
		return model.GradeResult{}, nil, err
	}
	return result.Clone(), preview.Clone(), nil
}

func (coordinator *Coordinator) Publish(runID string, trayID string) (model.GradeResult, error) {
	result, graded, err := coordinator.Preview(runID, trayID)
	if err != nil {
		return model.GradeResult{}, err
	}
	result.Published = true
	graded.Stage = model.StageComplete
	graded.UpdatedAt = time.Now().UTC()
	if err := coordinator.trays.Replace(graded); err != nil {
		return model.GradeResult{}, err
	}
	coordinator.state.PutResult(runID, result)
	return result.Clone(), nil
}

func (coordinator *Coordinator) Result(runID string) (model.GradeResult, bool) {
	return coordinator.state.Result(runID)
}
