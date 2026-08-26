package profile

import (
	"fmt"
	"time"
)

type Checkpoint struct {
	Index      int           `json:"index"`
	Paused     bool          `json:"paused"`
	Remaining  time.Duration `json:"remaining"`
	Steps      []Step        `json:"steps"`
	CapturedAt time.Time     `json:"captured_at"`
}

func (runner *Runner) Checkpoint() Checkpoint {
	index, paused, remaining := runner.state.Snapshot()
	return Checkpoint{Index: index, Paused: paused, Remaining: remaining, Steps: runner.state.Steps(), CapturedAt: time.Now().UTC()}
}

func Restore(checkpoint Checkpoint) (*Runner, error) {
	state, err := NewState(checkpoint.Steps)
	if err != nil {
		return nil, err
	}
	if checkpoint.Index < 0 || checkpoint.Index >= len(checkpoint.Steps) {
		return nil, fmt.Errorf("profile checkpoint index %d is invalid", checkpoint.Index)
	}
	state.index = checkpoint.Index
	state.remaining = checkpoint.Remaining
	state.paused = checkpoint.Paused
	runner := NewRunner(state, NewRestTimer())
	runner.remaining = checkpoint.Remaining
	if !checkpoint.Paused {
		runner.epoch = runner.timer.Start(checkpoint.Remaining)
	}
	return runner, nil
}
