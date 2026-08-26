package alarm

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

type EvaluationError struct {
	Reason string
}

func (evaluation *EvaluationError) Error() string {
	if evaluation == nil {
		return "<nil>"
	}
	return evaluation.Reason
}

type Limits struct {
	MaximumTemperature float64
	MinimumVoltage     float64
	MaximumVoltage     float64
}

type Evaluator struct {
	limits Limits
}

func NewEvaluator(limits Limits) *Evaluator {
	return &Evaluator{limits: limits}
}

func (evaluator *Evaluator) Evaluate(sample *model.Sample) (*model.AlarmEvent, error) {
	if sample == nil {
		return nil, &EvaluationError{Reason: "sample is nil"}
	}
	code := ""
	message := ""
	switch {
	case sample.Temperature > evaluator.limits.MaximumTemperature:
		code = "temperature-high"
		message = fmt.Sprintf("temperature %.2f exceeds %.2f", sample.Temperature, evaluator.limits.MaximumTemperature)
	case sample.Voltage < evaluator.limits.MinimumVoltage:
		code = "voltage-low"
		message = fmt.Sprintf("voltage %.3f below %.3f", sample.Voltage, evaluator.limits.MinimumVoltage)
	case sample.Voltage > evaluator.limits.MaximumVoltage:
		code = "voltage-high"
		message = fmt.Sprintf("voltage %.3f exceeds %.3f", sample.Voltage, evaluator.limits.MaximumVoltage)
	default:
		var evaluation *EvaluationError
		return nil, evaluation
	}
	event := &model.AlarmEvent{ID: uuid.NewString(), RunID: sample.RunID, TrayID: sample.TrayID, CellID: sample.CellID, Channel: sample.Channel, Level: model.AlarmCritical, Code: code, Message: message, OccurredAt: time.Now().UTC()}
	return event, nil
}
