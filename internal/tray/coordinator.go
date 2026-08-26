package tray

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

type Coordinator struct {
	registry *Registry
}

func NewCoordinator(registry *Registry) *Coordinator {
	return &Coordinator{registry: registry}
}

func (coordinator *Coordinator) Load(campaignID string, trayID string, channels int) (*model.Tray, error) {
	if campaignID == "" || trayID == "" || channels <= 0 {
		return nil, fmt.Errorf("campaign, tray and positive channel count are required")
	}
	now := time.Now().UTC()
	tray := &model.Tray{
		ID:         trayID,
		CampaignID: campaignID,
		RunID:      uuid.NewString(),
		Revision:   uuid.NewString(),
		Stage:      model.StageLoaded,
		Cells:      map[string]*model.Cell{},
		UpdatedAt:  now,
	}
	for channel := 1; channel <= channels; channel++ {
		id := fmt.Sprintf("%s-C%03d", trayID, channel)
		tray.Cells[id] = &model.Cell{ID: id, Channel: channel, Grade: "pending", Revision: uuid.NewString()}
	}
	if err := coordinator.registry.Register(tray); err != nil {
		return nil, err
	}
	return tray.Clone(), nil
}

func (coordinator *Coordinator) Advance(trayID string, stage model.Stage) (*model.Tray, error) {
	tray, ok := coordinator.registry.Lookup(trayID)
	if !ok {
		return nil, fmt.Errorf("tray %s not found", trayID)
	}
	if err := model.Advance(tray, stage, time.Now()); err != nil {
		return nil, err
	}
	tray.Revision = uuid.NewString()
	if err := coordinator.registry.Replace(tray); err != nil {
		return nil, err
	}
	return tray.Clone(), nil
}
