package alarm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

type IsolationSink interface {
	Latch(model.AlarmEvent) error
	Clear(model.AlarmEvent) error
}

type Integrator struct {
	alarms    *Coordinator
	isolation IsolationSink
}

func NewIntegrator(alarms *Coordinator, isolation IsolationSink) *Integrator {
	return &Integrator{alarms: alarms, isolation: isolation}
}

func (integrator *Integrator) Sample(sample *model.Sample) error {
	event, err := integrator.alarms.Inspect(sample)
	if err != nil || event == nil {
		return err
	}
	return integrator.isolation.Latch(*event)
}

func (integrator *Integrator) Clear(runID string, trayID string, cellID string, channel int, code string) error {
	if runID == "" || channel <= 0 {
		return fmt.Errorf("run and channel are required")
	}
	event := model.AlarmEvent{ID: uuid.NewString(), RunID: runID, TrayID: trayID, CellID: cellID, Channel: channel, Level: model.AlarmInfo, Code: code, Message: "alarm cleared", Cleared: true, OccurredAt: time.Now().UTC()}
	integrator.alarms.state.Apply(event)
	if err := integrator.alarms.recorder.Alarm(event); err != nil {
		return err
	}
	return integrator.isolation.Clear(event)
}
