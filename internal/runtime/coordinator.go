package runtime

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/journal"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/sample"
)

func (system *System) Operations() []model.Operation {
	return system.campaigns.List()
}

func (system *System) Equipment() []model.EquipmentState {
	return system.deviceGateway.Equipment()
}

func (system *System) Interlocks() []model.Interlock {
	return system.isolationFlow.Interlocks("")
}

func (system *System) Incidents() []model.Incident {
	system.mu.RLock()
	out := make([]model.Incident, 0, len(system.incidents))
	for _, incident := range system.incidents {
		out = append(out, incident)
	}
	system.mu.RUnlock()
	sort.Slice(out, func(i int, j int) bool { return out[i].OpenedAt.After(out[j].OpenedAt) })
	return out
}

func (system *System) LoadCampaign(campaignID string, trayID string, channels int) (model.Operation, error) {
	return system.campaigns.Load(campaignID, trayID, channels)
}

func (system *System) SetCurrent(runID string, channel int, value float64) error {
	return system.deviceGateway.SetCurrent(runID, channel, value)
}

func (system *System) InjectSample(runID string, trayID string, cellID string, channel int, sequence uint64, voltage float64, temperature float64) (*model.Sample, error) {
	raw := sample.Encode(channel, sequence, voltage, temperature, []byte(fmt.Sprintf("%s/%d", cellID, sequence)))
	framed := cycler.EncodeFrame(raw)
	values, err := system.samples.ReadDeviceStream(bytes.NewReader(framed), runID, trayID, func(number int) string {
		if number == channel {
			return cellID
		}
		return ""
	})
	if err != nil {
		return nil, err
	}
	if len(values) != 1 {
		return nil, fmt.Errorf("device stream produced %d samples", len(values))
	}
	if err := system.devices.RecordSequence(channel, sequence); err != nil {
		return nil, err
	}
	value, ok := system.sampleHistory.Latest(runID, cellID)
	if !ok {
		return nil, fmt.Errorf("sample was not retained")
	}
	return value.Clone(), nil
}

func (system *System) RaiseIncident(runID string, trayID string, summary string, severity model.AlarmLevel) model.Incident {
	incident := model.Incident{ID: uuid.NewString(), RunID: runID, TrayID: trayID, Summary: summary, Severity: severity, Open: true, OpenedAt: time.Now().UTC()}
	system.mu.Lock()
	system.incidents[incident.ID] = incident
	system.mu.Unlock()
	return incident
}

func (system *System) AcknowledgeIncident(id string) (model.Incident, error) {
	system.mu.Lock()
	defer system.mu.Unlock()
	incident, ok := system.incidents[id]
	if !ok {
		return model.Incident{}, fmt.Errorf("incident %s not found", id)
	}
	incident.Open = false
	system.incidents[id] = incident
	return incident, nil
}

func (system *System) RecentEvents(limit int) []journal.Event {
	return system.recorder.Recent("", limit)
}

func (system *System) Samples(runID string) []*model.Sample {
	return system.quality.Checkpoint(runID).Samples
}

func (system *System) Curve(trayID string) ([]model.Cell, error) {
	return system.traySamples.Curve(trayID)
}

func (system *System) RuntimeMetrics() map[string]any {
	latched := false
	for _, operation := range system.campaigns.List() {
		latched = latched || system.alarms.Latched(operation.RunID)
	}
	return map[string]any{
		"events":           system.recorder.EventCount(),
		"samples":          system.sampleHistory.Count(""),
		"subscribers":      system.sampleBus.SubscriberCount(),
		"active_campaigns": len(system.campaigns.Active()),
		"alarm_latched":    latched,
	}
}
