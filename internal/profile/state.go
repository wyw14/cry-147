package profile

import (
	"fmt"
	"sync"
	"time"
)

type Step struct {
	Name               string        `json:"name"`
	Duration           time.Duration `json:"duration"`
	MinimumTemperature float64       `json:"minimum_temperature"`
	MaximumTemperature float64       `json:"maximum_temperature"`
}

type State struct {
	mu        sync.RWMutex
	steps     []Step
	index     int
	paused    bool
	remaining time.Duration
}

func NewState(steps []Step) (*State, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("profile needs at least one step")
	}
	copySteps := append([]Step(nil), steps...)
	return &State{steps: copySteps, remaining: copySteps[0].Duration}, nil
}

func (state *State) Current() (Step, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	if state.index >= len(state.steps) {
		return Step{}, false
	}
	return state.steps[state.index], true
}

func (state *State) Advance() (Step, bool) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.index++
	state.paused = false
	if state.index >= len(state.steps) {
		state.remaining = 0
		return Step{}, false
	}
	state.remaining = state.steps[state.index].Duration
	return state.steps[state.index], true
}

func (state *State) Pause(remaining time.Duration) {
	state.mu.Lock()
	state.paused = true
	state.remaining = remaining
	state.mu.Unlock()
}

func (state *State) Resume() time.Duration {
	state.mu.Lock()
	state.paused = false
	remaining := state.remaining
	state.mu.Unlock()
	return remaining
}

func (state *State) Snapshot() (int, bool, time.Duration) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.index, state.paused, state.remaining
}

func (state *State) Steps() []Step {
	state.mu.RLock()
	defer state.mu.RUnlock()
	return append([]Step(nil), state.steps...)
}
