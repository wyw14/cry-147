package runtime

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/alarm"
	"github.com/wyw14/cry-147/internal/campaign"
	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/isolation"
	"github.com/wyw14/cry-147/internal/journal"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/profile"
	"github.com/wyw14/cry-147/internal/quality"
	"github.com/wyw14/cry-147/internal/thermal"
)

type Checkpoint struct {
	Revision   string                        `json:"revision"`
	Campaigns  campaign.Checkpoint           `json:"campaigns"`
	Trays      []*model.Tray                 `json:"trays"`
	Equipment  cycler.Recovery               `json:"equipment"`
	Alarms     alarm.Checkpoint              `json:"alarms"`
	Isolation  isolation.Checkpoint          `json:"isolation"`
	Quality    map[string]quality.Checkpoint `json:"quality"`
	Profiles   map[string]profile.Checkpoint `json:"profiles"`
	Thermal    thermal.Checkpoint            `json:"thermal"`
	Incidents  []model.Incident              `json:"incidents"`
	CapturedAt time.Time                     `json:"captured_at"`
}

func (system *System) Checkpoint() Checkpoint {
	qualityCheckpoints := map[string]quality.Checkpoint{}
	for _, operation := range system.campaigns.List() {
		qualityCheckpoints[operation.RunID] = system.quality.Checkpoint(operation.RunID)
	}
	profileCheckpoints := map[string]profile.Checkpoint{}
	system.mu.RLock()
	for runID, integrator := range system.profiles {
		profileCheckpoints[runID] = integrator.Checkpoint()
	}
	system.mu.RUnlock()
	return Checkpoint{
		Revision:   uuid.NewString(),
		Campaigns:  system.campaigns.Checkpoint(),
		Trays:      system.trays.List(),
		Equipment:  system.deviceGateway.Recovery(),
		Alarms:     system.alarms.Checkpoint(),
		Isolation:  system.isolation.Checkpoint(),
		Quality:    qualityCheckpoints,
		Profiles:   profileCheckpoints,
		Thermal:    system.thermalState.Checkpoint(),
		Incidents:  system.Incidents(),
		CapturedAt: time.Now().UTC(),
	}
}

func (system *System) RestoreCheckpoint(path string) (Checkpoint, error) {
	checkpoint, err := LoadCheckpoint(path)
	if err != nil {
		return Checkpoint{}, err
	}
	if err := system.trays.Restore(checkpoint.Trays); err != nil {
		return Checkpoint{}, fmt.Errorf("restore trays: %w", err)
	}
	if err := system.campaigns.Restore(checkpoint.Campaigns); err != nil {
		return Checkpoint{}, fmt.Errorf("restore campaigns: %w", err)
	}
	if err := system.devices.Restore(checkpoint.Equipment.States); err != nil {
		return Checkpoint{}, fmt.Errorf("restore equipment: %w", err)
	}
	if err := system.alarms.Restore(checkpoint.Alarms); err != nil {
		return Checkpoint{}, fmt.Errorf("restore alarms: %w", err)
	}
	if err := system.isolation.Restore(checkpoint.Isolation); err != nil {
		return Checkpoint{}, fmt.Errorf("restore isolation: %w", err)
	}
	for runID, value := range checkpoint.Quality {
		if runID != value.RunID {
			return Checkpoint{}, fmt.Errorf("quality checkpoint key %s does not match %s", runID, value.RunID)
		}
		if err := system.quality.Restore(value); err != nil {
			return Checkpoint{}, fmt.Errorf("restore quality %s: %w", runID, err)
		}
	}
	if err := system.thermalState.Restore(checkpoint.Thermal); err != nil {
		return Checkpoint{}, fmt.Errorf("restore thermal state: %w", err)
	}
	profiles := make(map[string]*profile.Integrator, len(checkpoint.Profiles))
	for runID, value := range checkpoint.Profiles {
		runner, err := profile.Restore(value)
		if err != nil {
			return Checkpoint{}, fmt.Errorf("restore profile %s: %w", runID, err)
		}
		profiles[runID] = profile.NewIntegrator(runner, system.campaigns)
	}
	incidents := make(map[string]model.Incident, len(checkpoint.Incidents))
	for _, incident := range checkpoint.Incidents {
		if incident.ID == "" {
			return Checkpoint{}, fmt.Errorf("recovered incident identity is incomplete")
		}
		incidents[incident.ID] = incident
	}
	system.mu.Lock()
	system.profiles = profiles
	system.incidents = incidents
	system.mu.Unlock()
	return checkpoint, nil
}

func (system *System) SaveCheckpoint() (string, error) {
	checkpoint := system.Checkpoint()
	path := filepath.Join(system.dataDir, "snapshot.json")
	if err := journal.WriteSnapshot(path, checkpoint.Revision, checkpoint); err != nil {
		return "", err
	}
	return path, nil
}

func LoadCheckpoint(path string) (Checkpoint, error) {
	var checkpoint Checkpoint
	envelope, err := journal.ReadSnapshot(path, &checkpoint)
	if err != nil {
		return Checkpoint{}, err
	}
	if checkpoint.Revision == "" || checkpoint.Revision != envelope.Revision {
		return Checkpoint{}, fmt.Errorf("snapshot revision mismatch")
	}
	return checkpoint, nil
}

func (system *System) RecoveryReport(runID string) (model.RecoveryReport, error) {
	recovered, err := journal.Rebuild(runID, system.recorder.Recent(runID, 10000))
	if err != nil {
		return model.RecoveryReport{}, err
	}
	report := system.traySamples.Summary(runID)
	if len(recovered.Operations) == 0 {
		return report, fmt.Errorf("run %s has no persisted operation", runID)
	}
	return report, nil
}
