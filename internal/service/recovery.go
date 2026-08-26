package service

import (
	"fmt"

	"github.com/wyw14/cry-147/internal/model"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
)

type RecoveryView struct {
	CheckpointPath string               `json:"checkpoint_path"`
	Report         model.RecoveryReport `json:"report"`
}

func (runtime *Runtime) SaveRecovery(runID string) (RecoveryView, error) {
	if runID == "" {
		return RecoveryView{}, fmt.Errorf("run_id is required")
	}
	path, err := runtime.system.SaveCheckpoint()
	if err != nil {
		return RecoveryView{}, err
	}
	report, err := runtime.system.RecoveryReport(runID)
	if err != nil {
		return RecoveryView{}, err
	}
	return RecoveryView{CheckpointPath: path, Report: report}, nil
}

func (runtime *Runtime) ValidateRecovery(path string) error {
	if path == "" {
		return fmt.Errorf("checkpoint path is required")
	}
	_, err := cellruntime.LoadCheckpoint(path)
	return err
}

func (runtime *Runtime) RestoreRecovery(path string) (cellruntime.Checkpoint, error) {
	if path == "" {
		return cellruntime.Checkpoint{}, fmt.Errorf("checkpoint path is required")
	}
	return runtime.system.RestoreCheckpoint(path)
}
