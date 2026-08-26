package journal

import (
	"encoding/json"
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type RecoveredRun struct {
	RunID      string              `json:"run_id"`
	Operations []model.Operation   `json:"operations"`
	Alarms     []model.AlarmEvent  `json:"alarms"`
	Samples    []*model.Sample     `json:"samples"`
	Grades     []model.GradeResult `json:"grades"`
	Rejected   int                 `json:"rejected"`
}

func Rebuild(runID string, events []Event) (RecoveredRun, error) {
	recovered := RecoveredRun{RunID: runID}
	for _, event := range events {
		if event.RunID != runID {
			continue
		}
		switch event.Kind {
		case "campaign.stage":
			var operation model.Operation
			if err := json.Unmarshal(event.Payload, &operation); err != nil {
				return recovered, fmt.Errorf("stage event %s: %w", event.ID, err)
			}
			recovered.Operations = append(recovered.Operations, operation)
		case "alarm.changed":
			var alarm model.AlarmEvent
			if err := json.Unmarshal(event.Payload, &alarm); err != nil {
				return recovered, fmt.Errorf("alarm event %s: %w", event.ID, err)
			}
			recovered.Alarms = append(recovered.Alarms, alarm)
		case "sample.captured":
			var sample model.Sample
			if err := json.Unmarshal(event.Payload, &sample); err != nil {
				return recovered, fmt.Errorf("sample event %s: %w", event.ID, err)
			}
			recovered.Samples = append(recovered.Samples, sample.Clone())
		case "quality.grade":
			var result model.GradeResult
			if err := json.Unmarshal(event.Payload, &result); err != nil {
				return recovered, fmt.Errorf("grade event %s: %w", event.ID, err)
			}
			recovered.Grades = append(recovered.Grades, result.Clone())
		default:
			recovered.Rejected++
		}
	}
	return recovered, nil
}
