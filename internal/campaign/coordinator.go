package campaign

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/tray"
)

type Recorder interface {
	Stage(model.Operation) error
}

type Coordinator struct {
	mu       sync.RWMutex
	trays    *tray.Coordinator
	recorder Recorder
	states   map[string]*State
}

func NewCoordinator(trays *tray.Coordinator, recorder Recorder) *Coordinator {
	return &Coordinator{trays: trays, recorder: recorder, states: map[string]*State{}}
}

func (coordinator *Coordinator) Load(campaignID string, trayID string, channels int) (model.Operation, error) {
	loaded, err := coordinator.trays.Load(campaignID, trayID, channels)
	if err != nil {
		return model.Operation{}, err
	}
	now := time.Now().UTC()
	operation := model.Operation{ID: uuid.NewString(), RunID: loaded.RunID, TrayID: loaded.ID, Stage: model.StageLoaded, StartedAt: now, UpdatedAt: now, Message: "tray loaded"}
	coordinator.mu.Lock()
	coordinator.states[operation.RunID] = NewState(operation)
	coordinator.mu.Unlock()
	if err := coordinator.recorder.Stage(operation); err != nil {
		return model.Operation{}, err
	}
	return operation, nil
}

func (coordinator *Coordinator) Advance(runID string, stage model.Stage, message string) (model.Operation, error) {
	coordinator.mu.RLock()
	state, ok := coordinator.states[runID]
	coordinator.mu.RUnlock()
	if !ok {
		return model.Operation{}, fmt.Errorf("run %s not found", runID)
	}
	if err := state.Advance(stage, message); err != nil {
		return model.Operation{}, err
	}
	operation := state.Snapshot()
	if _, err := coordinator.trays.Advance(operation.TrayID, stage); err != nil {
		return model.Operation{}, err
	}
	if err := coordinator.recorder.Stage(operation); err != nil {
		return model.Operation{}, err
	}
	return operation, nil
}

func (coordinator *Coordinator) List() []model.Operation {
	coordinator.mu.RLock()
	states := make([]*State, 0, len(coordinator.states))
	for _, state := range coordinator.states {
		states = append(states, state)
	}
	coordinator.mu.RUnlock()
	out := make([]model.Operation, 0, len(states))
	for _, state := range states {
		out = append(out, state.Snapshot())
	}
	sort.Slice(out, func(i int, j int) bool { return out[i].StartedAt.Before(out[j].StartedAt) })
	return out
}
