package tray

import (
	"fmt"
	"sort"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type Snapshot struct {
	Tray       *model.Tray `json:"tray"`
	CellOrder  []string    `json:"cell_order"`
	CapturedAt time.Time   `json:"captured_at"`
}

func Freeze(input *model.Tray) (Snapshot, error) {
	if err := model.ValidateTray(input); err != nil {
		return Snapshot{}, err
	}
	copyValue := *input
	copyTray := &copyValue
	order := make([]string, 0, len(copyTray.Cells))
	for id := range copyTray.Cells {
		order = append(order, id)
	}
	sort.Strings(order)
	return Snapshot{Tray: copyTray, CellOrder: order, CapturedAt: time.Now().UTC()}, nil
}

func (snapshot Snapshot) Cell(id string) (*model.Cell, error) {
	if snapshot.Tray == nil {
		return nil, fmt.Errorf("snapshot has no tray")
	}
	cell, ok := snapshot.Tray.Cells[id]
	if !ok {
		return nil, fmt.Errorf("cell %s not found", id)
	}
	return cell.Clone(), nil
}

func (snapshot Snapshot) Clone() Snapshot {
	out := snapshot
	copyTray := *snapshot.Tray
	out.Tray = &copyTray
	out.CellOrder = append([]string(nil), snapshot.CellOrder...)
	return out
}

func (snapshot Snapshot) Capacities() map[string]float64 {
	values := map[string]float64{}
	if snapshot.Tray == nil {
		return values
	}
	for id, cell := range snapshot.Tray.Cells {
		values[id] = cell.Capacity
	}
	return values
}
