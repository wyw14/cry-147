package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SnapshotEnvelope struct {
	Revision  string          `json:"revision"`
	WrittenAt time.Time       `json:"written_at"`
	Value     json.RawMessage `json:"value"`
}

func WriteSnapshot(path string, revision string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	envelope := SnapshotEnvelope{Revision: revision, WrittenAt: time.Now().UTC(), Value: payload}
	encoded, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, encoded, 0o644); err != nil {
		return err
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}
	return nil
}

func ReadSnapshot(path string, destination any) (SnapshotEnvelope, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SnapshotEnvelope{}, err
	}
	var envelope SnapshotEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return SnapshotEnvelope{}, err
	}
	if err := json.Unmarshal(envelope.Value, destination); err != nil {
		return SnapshotEnvelope{}, err
	}
	return envelope, nil
}
