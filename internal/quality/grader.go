package quality

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

type Thresholds struct {
	GradeA  float64
	GradeB  float64
	Minimum float64
}

type Grader struct {
	thresholds Thresholds
}

func NewGrader(thresholds Thresholds) *Grader {
	return &Grader{thresholds: thresholds}
}

func (grader *Grader) Grade(capacity float64) string {
	switch {
	case capacity >= grader.thresholds.GradeA:
		return "A"
	case capacity >= grader.thresholds.GradeB:
		return "B"
	case capacity >= grader.thresholds.Minimum:
		return "C"
	default:
		return "reject"
	}
}

func (grader *Grader) Calculate(tray *model.Tray, samples []*model.Sample) (model.GradeResult, error) {
	if tray == nil {
		return model.GradeResult{}, fmt.Errorf("tray is required")
	}
	capacityByCell := map[string]float64{}
	for _, sample := range samples {
		if sample == nil || sample.RunID != tray.RunID {
			continue
		}
		capacityByCell[sample.CellID] += math.Abs(sample.Voltage) / 1000
	}
	ids := make([]string, 0, len(tray.Cells))
	for id := range tray.Cells {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	counts := map[string]int{}
	total := 0.0
	for _, id := range ids {
		capacity := capacityByCell[id]
		cell := tray.Cells[id]
		cell.Capacity = capacity
		cell.Grade = grader.Grade(capacity)
		cell.Revision = uuid.NewString()
		counts[cell.Grade]++
		total += capacity
	}
	average := 0.0
	if len(ids) > 0 {
		average = total / float64(len(ids))
	}
	return model.GradeResult{TrayID: tray.ID, Revision: uuid.NewString(), Counts: counts, AverageCapacity: average, CreatedAt: time.Now().UTC()}, nil
}
