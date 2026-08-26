package model

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidTransition = errors.New("invalid campaign transition")

var stageOrder = map[Stage]int{
	StageLoaded:       1,
	StageConditioning: 2,
	StageResting:      3,
	StageGrading:      4,
	StageIsolating:    5,
	StageComplete:     6,
	StageFailed:       7,
}

func CanAdvance(from Stage, to Stage) bool {
	if from == to {
		return true
	}
	if to == StageFailed || to == StageIsolating {
		return from != StageComplete
	}
	allowed := map[Stage]Stage{
		StageLoaded:       StageConditioning,
		StageConditioning: StageResting,
		StageResting:      StageGrading,
		StageGrading:      StageComplete,
		StageIsolating:    StageGrading,
	}
	return allowed[from] == to
}

func Advance(tray *Tray, next Stage, now time.Time) error {
	if tray == nil {
		return errors.New("tray is nil")
	}
	if !CanAdvance(tray.Stage, next) {
		return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, tray.Stage, next)
	}
	tray.Stage = next
	tray.UpdatedAt = now.UTC()
	return nil
}

func StageRank(stage Stage) int {
	return stageOrder[stage]
}

func ActiveStage(stage Stage) bool {
	return stage != StageComplete && stage != StageFailed
}

func ProtectiveStage(stage Stage) bool {
	return stage == StageIsolating || stage == StageFailed
}

func TerminalStage(stage Stage) bool {
	return stage == StageComplete || stage == StageFailed
}

func StageLabel(stage Stage) string {
	labels := map[Stage]string{
		StageLoaded:       "Loaded",
		StageConditioning: "Conditioning",
		StageResting:      "Resting",
		StageGrading:      "Grading",
		StageIsolating:    "Isolating",
		StageComplete:     "Complete",
		StageFailed:       "Failed",
	}
	return labels[stage]
}
