package alarm_test

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-147/internal/alarm"
	"github.com/wyw14/cry-147/internal/model"
)

type alarmRecorder struct {
	events chan model.AlarmEvent
}

func (recorder *alarmRecorder) Alarm(event model.AlarmEvent) error {
	select {
	case recorder.events <- event:
	default:
	}
	return nil
}

func TestClosedAlarmStreamDoesNotEmitZeroClearEvents(t *testing.T) {
	recorder := &alarmRecorder{events: make(chan model.AlarmEvent, 4)}
	coordinator := alarm.NewCoordinator(alarm.NewEvaluator(alarm.Limits{MaximumTemperature: 50}), alarm.NewState(), recorder)
	stream := make(chan model.AlarmEvent)
	close(stream)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- coordinator.Consume(ctx, stream) }()
	select {
	case event := <-recorder.events:
		t.Fatalf("closed stream produced zero event %+v", event)
	case err := <-result:
		if err != nil {
			t.Fatalf("closed stream returned error %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("consumer did not terminate after stream close")
	}
}
