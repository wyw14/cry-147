package profile

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Runner struct {
	mu        sync.Mutex
	state     *State
	timer     *RestTimer
	epoch     uint64
	remaining time.Duration
}

func NewRunner(state *State, timer *RestTimer) *Runner {
	return &Runner{state: state, timer: timer}
}

func (runner *Runner) BeginRest() (uint64, error) {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	step, ok := runner.state.Current()
	if !ok {
		return 0, fmt.Errorf("profile is complete")
	}
	runner.remaining = step.Duration
	runner.epoch = runner.timer.Start(step.Duration)
	return runner.epoch, nil
}

func (runner *Runner) PauseRest() time.Duration {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	runner.remaining = runner.timer.Pause()
	runner.state.Pause(runner.remaining)
	return runner.remaining
}

func (runner *Runner) ResumeRest() uint64 {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	runner.remaining = runner.state.Resume()
	runner.epoch = runner.timer.Resume(runner.remaining)
	return runner.epoch
}

func (runner *Runner) Wait(ctx context.Context, temperature float64) (Step, error) {
	runner.mu.Lock()
	epoch := runner.epoch
	runner.mu.Unlock()
	if _, err := runner.timer.Wait(ctx, epoch); err != nil {
		return Step{}, err
	}
	step, ok := runner.state.Current()
	if !ok {
		return Step{}, fmt.Errorf("profile is complete")
	}
	if err := ValidateTemperature(step, temperature); err != nil {
		return Step{}, err
	}
	next, _ := runner.state.Advance()
	return next, nil
}

func (runner *Runner) Remaining() time.Duration {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	return runner.remaining
}
