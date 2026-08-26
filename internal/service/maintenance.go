package service

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
	"github.com/wyw14/cry-147/internal/thermal"
)

type EquipmentResultRequest struct {
	RunID   string `json:"run_id"`
	Channel int    `json:"channel"`
	Busy    bool   `json:"busy"`
	Message string `json:"message"`
}

type ClearIsolationRequest struct {
	RunID   string `json:"run_id"`
	TrayID  string `json:"tray_id"`
	CellID  string `json:"cell_id"`
	Channel int    `json:"channel"`
	Code    string `json:"code"`
}

func (runtime *Runtime) Samples(runID string) ([]*model.Sample, error) {
	if runID == "" {
		return nil, fmt.Errorf("run_id is required")
	}
	return runtime.system.Samples(runID), nil
}

func (runtime *Runtime) Curve(trayID string) ([]model.Cell, error) {
	if trayID == "" {
		return nil, fmt.Errorf("tray_id is required")
	}
	return runtime.system.Curve(trayID)
}

func (runtime *Runtime) PauseRest(runID string) (cellruntime.RestStatus, error) {
	return runtime.system.PauseRest(runID)
}

func (runtime *Runtime) ResumeRest(runID string) (cellruntime.RestStatus, error) {
	return runtime.system.ResumeRest(runID)
}

func (runtime *Runtime) RestStatus(runID string) (cellruntime.RestStatus, error) {
	return runtime.system.RestStatus(runID)
}

func (runtime *Runtime) CompleteRest(ctx context.Context, runID string, temperature float64) (model.Operation, error) {
	return runtime.system.CompleteRest(ctx, runID, temperature)
}

func (runtime *Runtime) EquipmentResult(request EquipmentResultRequest) (map[string]any, error) {
	if request.RunID == "" || request.Channel <= 0 {
		return nil, fmt.Errorf("run_id and channel are required")
	}
	return runtime.system.EquipmentResult(request.RunID, request.Channel, request.Busy, request.Message)
}

func (runtime *Runtime) ClearIsolation(request ClearIsolationRequest) error {
	if request.RunID == "" || request.TrayID == "" || request.CellID == "" || request.Channel <= 0 || request.Code == "" {
		return fmt.Errorf("interlock identity and code are required")
	}
	return runtime.system.ClearIsolation(request.RunID, request.TrayID, request.CellID, request.Channel, request.Code)
}

func (runtime *Runtime) GradeResult(runID string) (model.GradeResult, bool) {
	return runtime.system.GradeResult(runID)
}

func (runtime *Runtime) ThermalReading(zone string) (thermal.Reading, bool) {
	return runtime.system.ThermalReading(zone)
}

func (runtime *Runtime) WaitThermalStable(ctx context.Context, zone string, interval time.Duration) error {
	return runtime.system.WaitThermalStable(ctx, zone, interval)
}

func (runtime *Runtime) IsolatedChannels(trayID string) ([]int, error) {
	return runtime.system.IsolatedChannels(trayID)
}
