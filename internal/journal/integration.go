package journal

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type Recorder struct {
	store *Store
}

func NewRecorder(store *Store) *Recorder {
	return &Recorder{store: store}
}

func (recorder *Recorder) Sample(sample *model.Sample) error {
	if sample == nil {
		return fmt.Errorf("sample is required")
	}
	copySample := sample.Clone()
	_, err := recorder.store.Append(copySample.RunID, "sample.captured", copySample.ID, copySample)
	return err
}

func (recorder *Recorder) Alarm(event model.AlarmEvent) error {
	_, err := recorder.store.Append(event.RunID, "alarm.changed", event.ID, event)
	return err
}

func (recorder *Recorder) Stage(operation model.Operation) error {
	_, err := recorder.store.Append(operation.RunID, "campaign.stage", operation.ID, operation)
	return err
}

func (recorder *Recorder) Grade(runID string, result model.GradeResult) error {
	_, err := recorder.store.Append(runID, "quality.grade", result.TrayID, result.Clone())
	return err
}

func (recorder *Recorder) Recent(runID string, limit int) []Event {
	return recorder.store.List(runID, limit)
}

func (recorder *Recorder) EventCount() int {
	return recorder.store.Count("")
}
