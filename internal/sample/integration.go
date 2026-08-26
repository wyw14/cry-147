package sample

import (
	"fmt"
	"io"

	"github.com/wyw14/cry-147/internal/cycler"
	"github.com/wyw14/cry-147/internal/model"
)

type Pipeline struct {
	decoder     *Decoder
	coordinator *Coordinator
	history     *History
}

func NewPipeline(decoder *Decoder, coordinator *Coordinator, history *History) *Pipeline {
	return &Pipeline{decoder: decoder, coordinator: coordinator, history: history}
}

func (pipeline *Pipeline) ReadDeviceStream(reader io.Reader, runID string, trayID string, cellForChannel func(int) string) ([]*model.Sample, error) {
	frames, err := cycler.ReadAllFrames(reader, 64*1024)
	if err != nil {
		return nil, err
	}
	out := make([]*model.Sample, 0, len(frames))
	for _, frame := range frames {
		if len(frame) < 4 {
			return nil, fmt.Errorf("decoded frame has no channel")
		}
		probe, err := pipeline.decoder.Decode(runID, trayID, "", frame)
		if err != nil {
			return nil, err
		}
		probe.CellID = cellForChannel(probe.Channel)
		if err := pipeline.coordinator.PublishPooled(*probe); err != nil {
			return nil, err
		}
		if err := pipeline.history.Add(probe); err != nil {
			return nil, err
		}
		out = append(out, probe.Clone())
	}
	return out, nil
}
