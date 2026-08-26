package campaign

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type Checkpoint struct {
	Operations []model.Operation `json:"operations"`
	CapturedAt time.Time         `json:"captured_at"`
}

func (coordinator *Coordinator) Checkpoint() Checkpoint {
	return Checkpoint{Operations: coordinator.List(), CapturedAt: time.Now().UTC()}
}

func (coordinator *Coordinator) Restore(checkpoint Checkpoint) error {
	restored := make(map[string]*State, len(checkpoint.Operations))
	for _, operation := range checkpoint.Operations {
		if operation.RunID == "" || operation.TrayID == "" {
			return fmt.Errorf("recovered operation identity is incomplete")
		}
		if _, exists := restored[operation.RunID]; exists {
			return fmt.Errorf("duplicate recovered run %s", operation.RunID)
		}
		restored[operation.RunID] = NewState(operation)
	}
	coordinator.mu.Lock()
	coordinator.states = restored
	coordinator.mu.Unlock()
	return nil
}

func (coordinator *Coordinator) Active() []model.Operation {
	all := coordinator.List()
	out := make([]model.Operation, 0, len(all))
	for _, operation := range all {
		if model.ActiveStage(operation.Stage) {
			out = append(out, operation)
		}
	}
	return out
}
