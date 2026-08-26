package journal

import (
	"encoding/json"
	"sort"
)

func cloneEvent(event Event) Event {
	copyEvent := event
	copyEvent.Payload = append(json.RawMessage(nil), event.Payload...)
	return copyEvent
}

func (store *Store) List(runID string, limit int) []Event {
	store.mu.Lock()
	defer store.mu.Unlock()
	if limit <= 0 {
		limit = len(store.events)
	}
	out := make([]Event, 0, limit)
	for index := len(store.events) - 1; index >= 0 && len(out) < limit; index-- {
		event := store.events[index]
		if runID == "" || event.RunID == runID {
			out = append(out, cloneEvent(event))
		}
	}
	sort.Slice(out, func(i int, j int) bool { return out[i].OccurredAt.Before(out[j].OccurredAt) })
	return out
}

func (store *Store) Count(kind string) int {
	store.mu.Lock()
	defer store.mu.Unlock()
	count := 0
	for _, event := range store.events {
		if kind == "" || event.Kind == kind {
			count++
		}
	}
	return count
}
