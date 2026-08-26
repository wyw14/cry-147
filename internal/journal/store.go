package journal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         string          `json:"id"`
	RunID      string          `json:"run_id"`
	Kind       string          `json:"kind"`
	EntityID   string          `json:"entity_id"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurred_at"`
}

type Store struct {
	mu     sync.Mutex
	path   string
	events []Event
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	store := &Store{path: path}
	if err := store.load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (store *Store) load() error {
	file, err := os.Open(store.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("decode journal: %w", err)
		}
		store.events = append(store.events, event)
	}
	return scanner.Err()
}

func (store *Store) Append(runID string, kind string, entityID string, value any) (Event, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return Event{}, err
	}
	event := Event{ID: uuid.NewString(), RunID: runID, Kind: kind, EntityID: entityID, Payload: payload, OccurredAt: time.Now().UTC()}
	store.mu.Lock()
	defer store.mu.Unlock()
	file, err := os.OpenFile(store.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return Event{}, err
	}
	encoded, err := json.Marshal(event)
	if err == nil {
		_, err = file.Write(append(encoded, '\n'))
	}
	closeErr := file.Close()
	if err != nil {
		return Event{}, err
	}
	if closeErr != nil {
		return Event{}, closeErr
	}
	store.events = append(store.events, event)
	return event, nil
}

func (store *Store) Path() string {
	return store.path
}
