package service

import (
	"context"
	"time"

	"github.com/wyw14/cry-147/internal/model"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
	"github.com/wyw14/cry-147/internal/thermal"
)

type Runtime struct {
	system    *cellruntime.System
	startedAt time.Time
}

func NewRuntime(system *cellruntime.System) *Runtime {
	return &Runtime{system: system, startedAt: time.Now().UTC()}
}

func (runtime *Runtime) Health(ctx context.Context) map[string]any {
	err := runtime.system.Ready(ctx)
	status := "ok"
	detail := "ready"
	if err != nil {
		status = "error"
		detail = err.Error()
	}
	return map[string]any{"status": status, "detail": detail, "started_at": runtime.startedAt, "data_dir": runtime.system.DataDir(), "journal": runtime.system.JournalPath(), "metrics": runtime.system.RuntimeMetrics(), "recent_events": runtime.system.RecentEvents(5)}
}

func (runtime *Runtime) Operations() []model.Operation {
	return runtime.system.Operations()
}

func (runtime *Runtime) Equipment() []model.EquipmentState {
	return runtime.system.Equipment()
}

func (runtime *Runtime) Interlocks() []model.Interlock {
	return runtime.system.Interlocks()
}

func (runtime *Runtime) Incidents() []model.Incident {
	return runtime.system.Incidents()
}

func (runtime *Runtime) PollThermal(ctx context.Context, zone string) (thermal.Reading, error) {
	return runtime.system.PollThermal(ctx, zone)
}
