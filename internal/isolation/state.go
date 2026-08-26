package isolation

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

type State struct {
	mu         sync.RWMutex
	interlocks map[string]model.Interlock
}

func NewState() *State {
	return &State{interlocks: map[string]model.Interlock{}}
}

func interlockKey(runID string, channel int) string {
	return fmt.Sprintf("%s/%d", runID, channel)
}

func (state *State) Latch(event model.AlarmEvent) model.Interlock {
	state.mu.Lock()
	defer state.mu.Unlock()
	key := interlockKey(event.RunID, event.Channel)
	interlock := model.Interlock{ID: uuid.NewString(), RunID: event.RunID, Channel: event.Channel, Reason: event.Message, Latched: true, UpdatedAt: time.Now().UTC()}
	state.interlocks[key] = interlock
	return interlock
}

func (state *State) Clear(event model.AlarmEvent) (model.Interlock, bool) {
	state.mu.Lock()
	defer state.mu.Unlock()
	key := interlockKey(event.RunID, event.Channel)
	interlock, ok := state.interlocks[key]
	if !ok {
		return model.Interlock{}, false
	}
	interlock.Latched = false
	interlock.UpdatedAt = time.Now().UTC()
	state.interlocks[key] = interlock
	return interlock, true
}

func (state *State) List(runID string) []model.Interlock {
	state.mu.RLock()
	out := []model.Interlock{}
	for _, interlock := range state.interlocks {
		if runID == "" || interlock.RunID == runID {
			out = append(out, interlock)
		}
	}
	state.mu.RUnlock()
	sort.Slice(out, func(i int, j int) bool { return out[i].UpdatedAt.Before(out[j].UpdatedAt) })
	return out
}
