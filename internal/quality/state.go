package quality

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-147/internal/model"
)

type State struct {
	mu      sync.RWMutex
	history map[string][]*model.Sample
	results map[string]model.GradeResult
}

func NewState() *State {
	return &State{history: map[string][]*model.Sample{}, results: map[string]model.GradeResult{}}
}

func (state *State) Accept(sample *model.Sample) error {
	if sample == nil || sample.RunID == "" {
		return fmt.Errorf("sample is incomplete")
	}
	state.mu.Lock()
	state.history[sample.RunID] = append(state.history[sample.RunID], sample.Clone())
	state.mu.Unlock()
	return nil
}

func (state *State) Samples(runID string) []*model.Sample {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return model.CloneSamples(state.history[runID])
}

func (state *State) PutResult(runID string, result model.GradeResult) {
	state.mu.Lock()
	state.results[runID] = result.Clone()
	state.mu.Unlock()
}

func (state *State) Result(runID string) (model.GradeResult, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	result, ok := state.results[runID]
	return result.Clone(), ok
}

func (state *State) Clear(runID string) {
	state.mu.Lock()
	delete(state.history, runID)
	delete(state.results, runID)
	state.mu.Unlock()
}
