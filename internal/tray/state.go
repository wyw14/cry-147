package tray

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

func UpdateCellSample(tray *model.Tray, sample *model.Sample) error {
	if tray == nil || sample == nil {
		return fmt.Errorf("tray and sample are required")
	}
	if sample.RunID != tray.RunID || sample.TrayID != tray.ID {
		return fmt.Errorf("sample ownership does not match tray")
	}
	cell, ok := tray.Cells[sample.CellID]
	if !ok || cell.Channel != sample.Channel {
		return fmt.Errorf("cell %s does not match channel %d", sample.CellID, sample.Channel)
	}
	cell.Voltage = sample.Voltage
	cell.Temperature = sample.Temperature
	cell.Revision = uuid.NewString()
	tray.Revision = uuid.NewString()
	tray.UpdatedAt = time.Now().UTC()
	return nil
}

func IsolateCell(tray *model.Tray, cellID string) error {
	if tray == nil {
		return fmt.Errorf("tray is required")
	}
	cell, ok := tray.Cells[cellID]
	if !ok {
		return fmt.Errorf("cell %s not found", cellID)
	}
	cell.Isolated = true
	cell.Revision = uuid.NewString()
	tray.Stage = model.StageIsolating
	tray.Revision = uuid.NewString()
	tray.UpdatedAt = time.Now().UTC()
	return nil
}

func ClearIsolation(tray *model.Tray, cellID string) error {
	if tray == nil {
		return fmt.Errorf("tray is required")
	}
	cell, ok := tray.Cells[cellID]
	if !ok {
		return fmt.Errorf("cell %s not found", cellID)
	}
	cell.Isolated = false
	cell.Revision = uuid.NewString()
	tray.Revision = uuid.NewString()
	tray.UpdatedAt = time.Now().UTC()
	return nil
}

func IsolatedChannels(tray *model.Tray) []int {
	channels := []int{}
	if tray == nil {
		return channels
	}
	for _, cell := range tray.Cells {
		if cell.Isolated {
			channels = append(channels, cell.Channel)
		}
	}
	return channels
}
