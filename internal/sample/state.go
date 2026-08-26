package sample

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-147/internal/model"
)

type History struct {
	mu           sync.RWMutex
	byRun        map[string][]*model.Sample
	lastSequence map[string]uint64
}

func NewHistory() *History {
	return &History{byRun: map[string][]*model.Sample{}, lastSequence: map[string]uint64{}}
}

func (history *History) Add(sample *model.Sample) error {
	if sample == nil || sample.RunID == "" {
		return fmt.Errorf("sample run identity is required")
	}
	history.mu.Lock()
	defer history.mu.Unlock()
	key := sample.RunID + "/" + sample.CellID
	if sample.Sequence <= history.lastSequence[key] {
		return fmt.Errorf("sample sequence %d is not after %d", sample.Sequence, history.lastSequence[key])
	}
	history.lastSequence[key] = sample.Sequence
	history.byRun[sample.RunID] = append(history.byRun[sample.RunID], sample.Clone())
	return nil
}

func (history *History) Latest(runID string, cellID string) (*model.Sample, bool) {
	history.mu.RLock()
	defer history.mu.RUnlock()
	items := history.byRun[runID]
	for index := len(items) - 1; index >= 0; index-- {
		if items[index].CellID == cellID {
			return items[index].Clone(), true
		}
	}
	return nil, false
}

func (history *History) Count(runID string) int {
	history.mu.RLock()
	defer history.mu.RUnlock()
	return len(history.byRun[runID])
}
