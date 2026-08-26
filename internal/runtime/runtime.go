package runtime

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/wyw14/cry-147/internal/alarm"
	"github.com/wyw14/cry-147/internal/campaign"
	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/isolation"
	"github.com/wyw14/cry-147/internal/journal"
	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/profile"
	"github.com/wyw14/cry-147/internal/quality"
	"github.com/wyw14/cry-147/internal/sample"
	"github.com/wyw14/cry-147/internal/thermal"
	"github.com/wyw14/cry-147/internal/tray"
)

type System struct {
	mu              sync.RWMutex
	dataDir         string
	journal         *journal.Store
	recorder        *journal.Recorder
	trays           *tray.Registry
	trayCoordinator *tray.Coordinator
	traySamples     *tray.SampleIntegrator
	devices         *cycler.Coordinator
	deviceGateway   *cycler.DeviceGateway
	campaigns       *campaign.Coordinator
	equipment       *campaign.EquipmentIntegrator
	samples         *sample.Pipeline
	sampleBus       *sample.Coordinator
	sampleHistory   *sample.History
	quality         *quality.Coordinator
	qualityFlow     *quality.Integrator
	alarms          *alarm.Coordinator
	alarmFlow       *alarm.Integrator
	isolation       *isolation.Coordinator
	isolationFlow   *isolation.Integrator
	thermalState    *thermal.State
	thermalFlow     *thermal.Integrator
	profiles        map[string]*profile.Integrator
	incidents       map[string]model.Incident
	alarmStream     chan model.AlarmEvent
}

func New(dataDir string, thermalURL string) (*System, error) {
	store, err := journal.Open(filepath.Join(dataDir, "events.jsonl"))
	if err != nil {
		return nil, err
	}
	recorder := journal.NewRecorder(store)
	trayRegistry := tray.NewRegistry()
	trayCoordinator := tray.NewCoordinator(trayRegistry)
	traySamples := tray.NewSampleIntegrator(trayRegistry)
	devices := cycler.NewCoordinator(64, 8)
	campaigns := campaign.NewCoordinator(trayCoordinator, recorder)
	equipment := campaign.NewEquipmentIntegrator(campaigns, devices)
	sampleBus := sample.NewCoordinator()
	sampleHistory := sample.NewHistory()
	samplePipeline := sample.NewPipeline(sample.NewDecoder(64*1024), sampleBus, sampleHistory)
	qualityState := quality.NewState()
	qualityCoordinator := quality.NewCoordinator(qualityState, trayRegistry, quality.NewGrader(quality.Thresholds{GradeA: 3.2, GradeB: 2.8, Minimum: 2.2}))
	qualityFlow := quality.NewIntegrator(qualityCoordinator, traySamples, recorder)
	isolationState := isolation.NewState()
	isolationCoordinator := isolation.NewCoordinator(isolationState, isolation.NewDispatcher(devices))
	isolationFlow := isolation.NewIntegrator(isolationCoordinator, trayRegistry)
	alarmState := alarm.NewState()
	alarmState.ObserveTrays(trayRegistry)
	alarmCoordinator := alarm.NewCoordinator(alarm.NewEvaluator(alarm.Limits{MaximumTemperature: 48, MinimumVoltage: 2.2, MaximumVoltage: 4.25}), alarmState, recorder)
	alarmFlow := alarm.NewIntegrator(alarmCoordinator, isolationFlow)
	thermalState := thermal.NewState()
	thermalClient, err := thermal.NewClient(thermalURL, &http.Client{Timeout: 3 * time.Second})
	if err != nil {
		return nil, err
	}
	thermalFlow := thermal.NewIntegrator(thermalClient, thermalState, 0.5)
	alarmStream := make(chan model.AlarmEvent, 128)
	system := &System{
		dataDir:         dataDir,
		journal:         store,
		recorder:        recorder,
		trays:           trayRegistry,
		trayCoordinator: trayCoordinator,
		traySamples:     traySamples,
		devices:         devices,
		deviceGateway:   cycler.NewDeviceGateway(devices),
		campaigns:       campaigns,
		equipment:       equipment,
		samples:         samplePipeline,
		sampleBus:       sampleBus,
		sampleHistory:   sampleHistory,
		quality:         qualityCoordinator,
		qualityFlow:     qualityFlow,
		alarms:          alarmCoordinator,
		alarmFlow:       alarmFlow,
		isolation:       isolationCoordinator,
		isolationFlow:   isolationFlow,
		thermalState:    thermalState,
		thermalFlow:     thermalFlow,
		profiles:        map[string]*profile.Integrator{},
		incidents:       map[string]model.Incident{},
		alarmStream:     alarmStream,
	}
	sampleBus.Subscribe("quality", qualityFlow.Sample)
	sampleBus.Subscribe("alarm", alarmFlow.Sample)
	go func() {
		_ = alarmCoordinator.Consume(context.Background(), alarmStream)
	}()
	if err := system.bootstrap(); err != nil {
		return nil, err
	}
	return system, nil
}

func (system *System) bootstrap() error {
	if len(system.campaigns.List()) > 0 {
		return nil
	}
	operation, err := system.campaigns.Load("CF-DEMO", "TRAY-001", 8)
	if err != nil {
		return err
	}
	if _, err := system.campaigns.Advance(operation.RunID, model.StageConditioning, "formation channels enabled"); err != nil {
		return err
	}
	for channel := 1; channel <= 8; channel++ {
		if err := system.deviceGateway.SetCurrent(operation.RunID, channel, 1.25); err != nil {
			return err
		}
	}
	if _, err := system.equipment.ApplyResult(operation.RunID, nil); err != nil {
		return err
	}
	return system.thermalFlow.SetTarget("rest-a", 25)
}

func (system *System) DataDir() string {
	return system.dataDir
}

func (system *System) JournalPath() string {
	return system.journal.Path()
}

func (system *System) Ready(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if system.journal == nil || system.trays == nil || system.devices == nil {
		return fmt.Errorf("runtime components are incomplete")
	}
	if len(system.campaigns.List()) == 0 {
		return fmt.Errorf("runtime has no operations")
	}
	lease, err := system.deviceGateway.Connect(ctx, func(context.Context) error { return nil })
	if err != nil {
		return err
	}
	if lease.ID() == "" {
		lease.Close()
		return fmt.Errorf("cycler admission returned an empty lease")
	}
	lease.Close()
	used, active := system.devices.AdmissionUsage()
	if used != 0 || active != 0 {
		return fmt.Errorf("cycler admission did not settle: used=%d active=%d", used, active)
	}
	for _, operation := range system.campaigns.List() {
		_ = model.StageRank(operation.Stage)
		_ = model.StageLabel(operation.Stage)
		_ = model.ProtectiveStage(operation.Stage)
		_ = model.TerminalStage(operation.Stage)
	}
	_ = system.alarms.Active("")
	_ = system.alarms.Recent(1)
	_ = system.sampleBus.SubscriberCount()
	return nil
}
