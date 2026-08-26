package quality

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type Checkpoint struct {
	RunID      string             `json:"run_id"`
	Samples    []*model.Sample    `json:"samples"`
	Result     *model.GradeResult `json:"result,omitempty"`
	CapturedAt time.Time          `json:"captured_at"`
}

func (coordinator *Coordinator) Checkpoint(runID string) Checkpoint {
	checkpoint := Checkpoint{RunID: runID, Samples: coordinator.state.Samples(runID), CapturedAt: time.Now().UTC()}
	if result, ok := coordinator.state.Result(runID); ok {
		copyResult := result.Clone()
		checkpoint.Result = &copyResult
	}
	return checkpoint
}

func (coordinator *Coordinator) Restore(checkpoint Checkpoint) error {
	if checkpoint.RunID == "" {
		return fmt.Errorf("quality checkpoint run is required")
	}
	coordinator.state.Clear(checkpoint.RunID)
	for _, sample := range checkpoint.Samples {
		if sample.RunID != checkpoint.RunID {
			return fmt.Errorf("sample %s belongs to another run", sample.ID)
		}
		if err := coordinator.state.Accept(sample); err != nil {
			return err
		}
	}
	if checkpoint.Result != nil {
		coordinator.state.PutResult(checkpoint.RunID, checkpoint.Result.Clone())
	}
	return nil
}
