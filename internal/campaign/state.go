package campaign

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type State struct {
	mu         sync.RWMutex
	operation  model.Operation
	retryCount int
	lastError  string
}

func NewState(operation model.Operation) *State {
	return &State{operation: operation}
}

func (state *State) Snapshot() model.Operation {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.operation
}

func (state *State) Advance(stage model.Stage, message string) error {
	state.mu.Lock()
	defer state.mu.Unlock()
	if !model.CanAdvance(state.operation.Stage, stage) {
		return fmt.Errorf("%w: %s to %s", model.ErrInvalidTransition, state.operation.Stage, stage)
	}
	state.operation.Stage = stage
	state.operation.Message = message
	state.operation.UpdatedAt = time.Now().UTC()
	return nil
}

func (state *State) DeviceResult(err error, retryable error) bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if err == nil {
		state.lastError = ""
		return false
	}
	state.lastError = err.Error()
	if errors.Is(err, retryable) {
		state.retryCount++
		state.operation.UpdatedAt = time.Now().UTC()
		return true
	}
	state.operation.Stage = model.StageFailed
	state.operation.UpdatedAt = time.Now().UTC()
	return false
}

func (state *State) RetryCount() int {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.retryCount
}
