package thermal

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Reading struct {
	Zone        string    `json:"zone"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	RecordedAt  time.Time `json:"recorded_at"`
}

type State struct {
	mu       sync.RWMutex
	readings map[string]Reading
	targets  map[string]float64
}

func NewState() *State {
	return &State{readings: map[string]Reading{}, targets: map[string]float64{}}
}

func (state *State) SetTarget(zone string, target float64) error {
	if zone == "" || target < -20 || target > 100 {
		return fmt.Errorf("invalid thermal target")
	}
	state.mu.Lock()
	state.targets[zone] = target
	state.mu.Unlock()
	return nil
}

func (state *State) Apply(reading Reading) error {
	if reading.Zone == "" || reading.RecordedAt.IsZero() {
		return fmt.Errorf("thermal reading is incomplete")
	}
	state.mu.Lock()
	state.readings[reading.Zone] = reading
	state.mu.Unlock()
	return nil
}

func (state *State) Reading(zone string) (Reading, bool) {
	state.mu.RLock()
	defer state.mu.RUnlock()
	reading, ok := state.readings[zone]
	return reading, ok
}

func (state *State) Stable(zone string, tolerance float64) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()
	reading, readingOK := state.readings[zone]
	target, targetOK := state.targets[zone]
	return readingOK && targetOK && math.Abs(reading.Temperature-target) <= tolerance
}
