package profile

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type expiration struct {
	epoch uint64
	at    time.Time
}

type RestTimer struct {
	mu       sync.Mutex
	epoch    uint64
	timer    *time.Timer
	events   chan expiration
	deadline time.Time
}

func NewRestTimer() *RestTimer {
	return &RestTimer{events: make(chan expiration, 8)}
}

func (timer *RestTimer) Start(duration time.Duration) uint64 {
	timer.mu.Lock()
	timer.stopLocked()
	timer.epoch++
	epoch := timer.epoch
	timer.deadline = time.Now().Add(duration)
	timer.timer = time.AfterFunc(duration, func() {
		timer.events <- expiration{epoch: epoch, at: time.Now().UTC()}
	})
	timer.mu.Unlock()
	return epoch
}

func (timer *RestTimer) stopLocked() {
	if timer.timer != nil {
		timer.timer.Stop()
		timer.timer = nil
	}
}

func (timer *RestTimer) Pause() time.Duration {
	timer.mu.Lock()
	defer timer.mu.Unlock()
	timer.stopLocked()
	remaining := time.Until(timer.deadline)
	if remaining < 0 {
		remaining = 0
	}
	timer.epoch++
	return remaining
}

func (timer *RestTimer) Resume(duration time.Duration) uint64 {
	return timer.Start(duration)
}

func (timer *RestTimer) Wait(ctx context.Context, epoch uint64) (time.Time, error) {
	for {
		select {
		case <-ctx.Done():
			return time.Time{}, ctx.Err()
		case event := <-timer.events:
			if event.epoch != epoch {
				continue
			}
			return event.at, nil
		}
	}
}

func ValidateTemperature(step Step, temperature float64) error {
	if temperature < step.MinimumTemperature || temperature > step.MaximumTemperature {
		return fmt.Errorf("temperature %.2f outside %.2f..%.2f", temperature, step.MinimumTemperature, step.MaximumTemperature)
	}
	return nil
}
