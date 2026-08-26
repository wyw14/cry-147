package service

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
)

type SampleRequest struct {
	RunID       string  `json:"run_id"`
	TrayID      string  `json:"tray_id"`
	CellID      string  `json:"cell_id"`
	Channel     int     `json:"channel"`
	Sequence    uint64  `json:"sequence"`
	Voltage     float64 `json:"voltage"`
	Temperature float64 `json:"temperature"`
}

type IsolationRequest struct {
	RunID   string `json:"run_id"`
	TrayID  string `json:"tray_id"`
	CellID  string `json:"cell_id"`
	Channel int    `json:"channel"`
	Reason  string `json:"reason"`
}

func (runtime *Runtime) Sample(request SampleRequest) (*model.Sample, error) {
	if request.RunID == "" || request.TrayID == "" || request.CellID == "" || request.Channel <= 0 || request.Sequence == 0 {
		return nil, fmt.Errorf("sample identity and positive sequence are required")
	}
	return runtime.system.InjectSample(request.RunID, request.TrayID, request.CellID, request.Channel, request.Sequence, request.Voltage, request.Temperature)
}

func (runtime *Runtime) Isolate(request IsolationRequest) error {
	if request.RunID == "" || request.TrayID == "" || request.CellID == "" || request.Channel <= 0 || request.Reason == "" {
		return fmt.Errorf("isolation identity and reason are required")
	}
	return runtime.system.Isolate(request.RunID, request.TrayID, request.CellID, request.Channel, request.Reason)
}

func (runtime *Runtime) PreviewGrade(runID string, trayID string) (model.GradeResult, *model.Tray, error) {
	if runID == "" || trayID == "" {
		return model.GradeResult{}, nil, fmt.Errorf("run_id and tray_id are required")
	}
	return runtime.system.PreviewGrade(runID, trayID)
}

func (runtime *Runtime) PublishGrade(runID string, trayID string) (model.GradeResult, error) {
	if runID == "" || trayID == "" {
		return model.GradeResult{}, fmt.Errorf("run_id and tray_id are required")
	}
	return runtime.system.PublishGrade(runID, trayID)
}
