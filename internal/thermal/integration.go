package thermal

import (
	"context"
	"fmt"
	"time"
)

type Integrator struct {
	client    *Client
	state     *State
	tolerance float64
}

func NewIntegrator(client *Client, state *State, tolerance float64) *Integrator {
	return &Integrator{client: client, state: state, tolerance: tolerance}
}

func (integrator *Integrator) Poll(ctx context.Context, zone string) (Reading, error) {
	reading, err := integrator.client.Reading(ctx, zone)
	if err != nil {
		return Reading{}, err
	}
	if err := integrator.state.Apply(reading); err != nil {
		return Reading{}, err
	}
	return reading, nil
}

func (integrator *Integrator) WaitStable(ctx context.Context, zone string, interval time.Duration) error {
	if interval <= 0 {
		return fmt.Errorf("poll interval must be positive")
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if integrator.state.Stable(zone, integrator.tolerance) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := integrator.Poll(ctx, zone); err != nil {
				return err
			}
		}
	}
}

func (integrator *Integrator) SetTarget(zone string, target float64) error {
	return integrator.state.SetTarget(zone, target)
}
