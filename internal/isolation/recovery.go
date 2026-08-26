package isolation

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type Checkpoint struct {
	Interlocks []model.Interlock `json:"interlocks"`
	CapturedAt time.Time         `json:"captured_at"`
}

func (coordinator *Coordinator) Checkpoint() Checkpoint {
	return Checkpoint{Interlocks: coordinator.List(""), CapturedAt: time.Now().UTC()}
}

func (coordinator *Coordinator) Restore(checkpoint Checkpoint) error {
	restored := make(map[string]model.Interlock, len(checkpoint.Interlocks))
	for _, interlock := range checkpoint.Interlocks {
		if interlock.ID == "" || interlock.RunID == "" || interlock.Channel <= 0 {
			return fmt.Errorf("interlock recovery record is incomplete")
		}
		key := interlockKey(interlock.RunID, interlock.Channel)
		if _, exists := restored[key]; exists {
			return fmt.Errorf("duplicate recovered interlock %s", key)
		}
		restored[key] = interlock
	}
	coordinator.state.mu.Lock()
	coordinator.state.interlocks = restored
	coordinator.state.mu.Unlock()
	return nil
}
