package model

import "time"

type Stage string

const (
	StageLoaded       Stage = "loaded"
	StageConditioning Stage = "conditioning"
	StageResting      Stage = "resting"
	StageGrading      Stage = "grading"
	StageIsolating    Stage = "isolating"
	StageComplete     Stage = "complete"
	StageFailed       Stage = "failed"
)

type Cell struct {
	ID          string  `json:"id"`
	Channel     int     `json:"channel"`
	Voltage     float64 `json:"voltage"`
	Temperature float64 `json:"temperature"`
	Capacity    float64 `json:"capacity"`
	Grade       string  `json:"grade"`
	Isolated    bool    `json:"isolated"`
	Revision    string  `json:"revision"`
}

type Tray struct {
	ID         string           `json:"id"`
	CampaignID string           `json:"campaign_id"`
	RunID      string           `json:"run_id"`
	Revision   string           `json:"revision"`
	Stage      Stage            `json:"stage"`
	Cells      map[string]*Cell `json:"cells"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type Sample struct {
	ID          string    `json:"id"`
	RunID       string    `json:"run_id"`
	TrayID      string    `json:"tray_id"`
	CellID      string    `json:"cell_id"`
	Channel     int       `json:"channel"`
	Sequence    uint64    `json:"sequence"`
	Voltage     float64   `json:"voltage"`
	Temperature float64   `json:"temperature"`
	Payload     []byte    `json:"payload"`
	CapturedAt  time.Time `json:"captured_at"`
}

type AlarmLevel string

const (
	AlarmInfo     AlarmLevel = "info"
	AlarmWarning  AlarmLevel = "warning"
	AlarmCritical AlarmLevel = "critical"
)

type AlarmEvent struct {
	ID         string     `json:"id"`
	RunID      string     `json:"run_id"`
	TrayID     string     `json:"tray_id"`
	CellID     string     `json:"cell_id"`
	Channel    int        `json:"channel"`
	Level      AlarmLevel `json:"level"`
	Code       string     `json:"code"`
	Message    string     `json:"message"`
	Cleared    bool       `json:"cleared"`
	OccurredAt time.Time  `json:"occurred_at"`
}

type CommandKind string

const (
	CommandSetCurrent CommandKind = "set_current"
	CommandSetVoltage CommandKind = "set_voltage"
	CommandPause      CommandKind = "pause"
	CommandResume     CommandKind = "resume"
	CommandStop       CommandKind = "stop"
)

type ChannelCommand struct {
	ID        string      `json:"id"`
	RunID     string      `json:"run_id"`
	Channel   int         `json:"channel"`
	Kind      CommandKind `json:"kind"`
	Value     float64     `json:"value"`
	Priority  int         `json:"priority"`
	CreatedAt time.Time   `json:"created_at"`
}

type GradeResult struct {
	TrayID          string         `json:"tray_id"`
	Revision        string         `json:"revision"`
	Counts          map[string]int `json:"counts"`
	AverageCapacity float64        `json:"average_capacity"`
	Published       bool           `json:"published"`
	CreatedAt       time.Time      `json:"created_at"`
}

type EquipmentState struct {
	Channel      int       `json:"channel"`
	SessionID    string    `json:"session_id"`
	Connected    bool      `json:"connected"`
	Protected    bool      `json:"protected"`
	Step         string    `json:"step"`
	SetCurrent   float64   `json:"set_current"`
	SetVoltage   float64   `json:"set_voltage"`
	LastSequence uint64    `json:"last_sequence"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Operation struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	TrayID    string    `json:"tray_id"`
	Stage     Stage     `json:"stage"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Message   string    `json:"message"`
}

type Interlock struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	Channel   int       `json:"channel"`
	Reason    string    `json:"reason"`
	Latched   bool      `json:"latched"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Incident struct {
	ID       string     `json:"id"`
	RunID    string     `json:"run_id"`
	TrayID   string     `json:"tray_id"`
	Summary  string     `json:"summary"`
	Severity AlarmLevel `json:"severity"`
	Open     bool       `json:"open"`
	OpenedAt time.Time  `json:"opened_at"`
}
