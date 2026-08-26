package model

import (
	"fmt"
	"sort"
	"time"
)

type RecoveryReport struct {
	RunID         string        `json:"run_id"`
	TrayCount     int           `json:"tray_count"`
	CellCount     int           `json:"cell_count"`
	IsolatedCount int           `json:"isolated_count"`
	LatestUpdate  time.Time     `json:"latest_update"`
	Stages        map[Stage]int `json:"stages"`
}

func BuildRecoveryReport(runID string, trays []*Tray) RecoveryReport {
	report := RecoveryReport{RunID: runID, Stages: map[Stage]int{}}
	for _, tray := range trays {
		if tray == nil || tray.RunID != runID {
			continue
		}
		report.TrayCount++
		report.Stages[tray.Stage]++
		if tray.UpdatedAt.After(report.LatestUpdate) {
			report.LatestUpdate = tray.UpdatedAt
		}
		for _, cell := range tray.Cells {
			report.CellCount++
			if cell != nil && cell.Isolated {
				report.IsolatedCount++
			}
		}
	}
	return report
}

func SortedCellIDs(tray *Tray) []string {
	if tray == nil {
		return nil
	}
	ids := make([]string, 0, len(tray.Cells))
	for id := range tray.Cells {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func ValidateTray(tray *Tray) error {
	if tray == nil {
		return fmt.Errorf("tray is required")
	}
	if tray.ID == "" || tray.RunID == "" || tray.Revision == "" {
		return fmt.Errorf("tray identity is incomplete")
	}
	if len(tray.Cells) == 0 {
		return fmt.Errorf("tray %s has no cells", tray.ID)
	}
	channels := map[int]string{}
	for id, cell := range tray.Cells {
		if cell == nil || cell.ID != id {
			return fmt.Errorf("tray %s has inconsistent cell %s", tray.ID, id)
		}
		if previous, ok := channels[cell.Channel]; ok {
			return fmt.Errorf("channel %d belongs to %s and %s", cell.Channel, previous, id)
		}
		channels[cell.Channel] = id
	}
	return nil
}
