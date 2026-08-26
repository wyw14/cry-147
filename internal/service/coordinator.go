package service

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-147/internal/model"
)

type CampaignRequest struct {
	CampaignID string `json:"campaign_id"`
	TrayID     string `json:"tray_id"`
	Channels   int    `json:"channels"`
}

type CurrentRequest struct {
	RunID   string  `json:"run_id"`
	Channel int     `json:"channel"`
	Amps    float64 `json:"amps"`
}

func (runtime *Runtime) LoadCampaign(request CampaignRequest) (model.Operation, error) {
	if request.CampaignID == "" || request.TrayID == "" || request.Channels < 1 || request.Channels > 64 {
		return model.Operation{}, fmt.Errorf("campaign_id, tray_id and channels 1..64 are required")
	}
	return runtime.system.LoadCampaign(request.CampaignID, request.TrayID, request.Channels)
}

func (runtime *Runtime) SetCurrent(request CurrentRequest) error {
	if request.RunID == "" || request.Channel <= 0 || request.Amps < 0 || request.Amps > 10 {
		return fmt.Errorf("run_id, channel and amps 0..10 are required")
	}
	return runtime.system.SetCurrent(request.RunID, request.Channel, request.Amps)
}

func (runtime *Runtime) StartRest(runID string, seconds int) error {
	if seconds < 1 || seconds > 3600 {
		return fmt.Errorf("rest seconds must be 1..3600")
	}
	return runtime.system.StartRest(runID, time.Duration(seconds)*time.Second, 20, 30)
}

func (runtime *Runtime) AcknowledgeIncident(id string) (model.Incident, error) {
	return runtime.system.AcknowledgeIncident(id)
}
