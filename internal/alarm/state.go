package alarm

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/tray"
)

type State struct {
	mu      sync.RWMutex
	active  map[string]model.AlarmEvent
	history []model.AlarmEvent
}

func NewState() *State {
	return &State{active: map[string]model.AlarmEvent{}}
}

func alarmKey(event model.AlarmEvent) string {
	return event.RunID + "/" + event.CellID + "/" + event.Code
}

func (state *State) Apply(event model.AlarmEvent) {
	state.mu.Lock()
	key := alarmKey(event)
	if event.Cleared {
		delete(state.active, key)
	} else {
		state.active[key] = event
	}
	state.history = append(state.history, event)
	state.mu.Unlock()
}

func (state *State) Active(runID string) []model.AlarmEvent {
	state.mu.RLock()
	out := []model.AlarmEvent{}
	for _, event := range state.active {
		if runID == "" || event.RunID == runID {
			out = append(out, event)
		}
	}
	state.mu.RUnlock()
	sort.Slice(out, func(i int, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out
}

func (state *State) History(limit int) []model.AlarmEvent {
	state.mu.RLock()
	defer state.mu.RUnlock()
	if limit <= 0 || limit > len(state.history) {
		limit = len(state.history)
	}
	start := len(state.history) - limit
	return append([]model.AlarmEvent(nil), state.history[start:]...)
}

func (state *State) Latched(runID string) bool {
	return len(state.Active(runID)) > 0
}

func (state *State) ObserveTrays(registry *tray.Registry) func() {
	return registry.Subscribe("alarm-state", func(snapshot model.Tray) {
		for _, cell := range snapshot.Cells {
			if !cell.Isolated {
				continue
			}
			state.Apply(model.AlarmEvent{ID: uuid.NewString(), RunID: snapshot.RunID, TrayID: snapshot.ID, CellID: cell.ID, Channel: cell.Channel, Level: model.AlarmCritical, Code: "tray-isolated", Message: "tray isolation observed", OccurredAt: time.Now().UTC()})
		}
	})
}
