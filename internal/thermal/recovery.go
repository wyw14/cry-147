package thermal

import (
	"fmt"
	"time"
)

type Checkpoint struct {
	Readings   map[string]Reading `json:"readings"`
	Targets    map[string]float64 `json:"targets"`
	CapturedAt time.Time          `json:"captured_at"`
}

func (state *State) Checkpoint() Checkpoint {
	state.mu.RLock()
	defer state.mu.RUnlock()
	readings := make(map[string]Reading, len(state.readings))
	for zone, reading := range state.readings {
		readings[zone] = reading
	}
	targets := make(map[string]float64, len(state.targets))
	for zone, target := range state.targets {
		targets[zone] = target
	}
	return Checkpoint{Readings: readings, Targets: targets, CapturedAt: time.Now().UTC()}
}

func (state *State) Restore(checkpoint Checkpoint) error {
	restored := NewState()
	for zone, target := range checkpoint.Targets {
		if err := restored.SetTarget(zone, target); err != nil {
			return fmt.Errorf("restore target %s: %w", zone, err)
		}
	}
	for _, reading := range checkpoint.Readings {
		if err := restored.Apply(reading); err != nil {
			return fmt.Errorf("restore reading %s: %w", reading.Zone, err)
		}
	}
	state.mu.Lock()
	state.readings = restored.readings
	state.targets = restored.targets
	state.mu.Unlock()
	return nil
}
