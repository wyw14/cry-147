package alarm

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type Checkpoint struct {
	Active     []model.AlarmEvent `json:"active"`
	History    []model.AlarmEvent `json:"history"`
	CapturedAt time.Time          `json:"captured_at"`
}

func (coordinator *Coordinator) Checkpoint() Checkpoint {
	return Checkpoint{Active: coordinator.state.Active(""), History: coordinator.state.History(0), CapturedAt: time.Now().UTC()}
}

func (coordinator *Coordinator) Restore(checkpoint Checkpoint) error {
	history := make([]model.AlarmEvent, 0, len(checkpoint.History))
	for _, event := range checkpoint.History {
		if event.ID == "" || event.RunID == "" {
			return fmt.Errorf("alarm recovery event is incomplete")
		}
		history = append(history, event)
	}
	active := make(map[string]model.AlarmEvent, len(checkpoint.Active))
	for _, event := range checkpoint.Active {
		if event.Cleared {
			return fmt.Errorf("active alarm %s is marked cleared", event.ID)
		}
		if event.ID == "" || event.RunID == "" {
			return fmt.Errorf("active alarm recovery event is incomplete")
		}
		active[alarmKey(event)] = event
	}
	coordinator.state.mu.Lock()
	coordinator.state.active = active
	coordinator.state.history = history
	coordinator.state.mu.Unlock()
	return nil
}
