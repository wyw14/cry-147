package isolation_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/isolation"
	"github.com/wyw14/cry-147/internal/model"
)

func TestIsolationStopPreemptsQueuedSetpoints(t *testing.T) {
	devices := cycler.NewCoordinator(2, 1)
	errors := make(chan error, 16)
	var group sync.WaitGroup
	for index := 0; index < 16; index++ {
		group.Add(1)
		go func(value int) {
			defer group.Done()
			_, err := devices.Schedule("run-2", 1, model.CommandSetCurrent, float64(value)/10)
			errors <- err
		}(index)
	}
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	controller := isolation.NewCoordinator(isolation.NewState(), isolation.NewDispatcher(devices))
	event := model.AlarmEvent{ID: "alarm-2", RunID: "run-2", TrayID: "tray-2", CellID: "cell-2", Channel: 1, Level: model.AlarmCritical, Code: "temperature-high", Message: fmt.Sprintf("latched at %s", time.Now().UTC())}
	if err := controller.Latch(event); err != nil {
		t.Fatal(err)
	}
	states := devices.States()
	if len(states) < 1 || !states[0].Protected || states[0].Step != "stopped" {
		t.Fatalf("channel was not stopped first: %+v", states)
	}
	if !controller.Protected("run-2", 1) {
		t.Fatal("interlock was not retained")
	}
}
