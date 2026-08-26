package sample

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-147/internal/model"
)

const frameHeaderSize = 28

type Decoder struct {
	mu      sync.Mutex
	scratch []byte
}

func NewDecoder(maxPayload int) *Decoder {
	if maxPayload < 64 {
		maxPayload = 4096
	}
	return &Decoder{scratch: make([]byte, maxPayload)}
}

func (decoder *Decoder) Decode(runID string, trayID string, cellID string, frame []byte) (*model.Sample, error) {
	decoder.mu.Lock()
	defer decoder.mu.Unlock()
	if len(frame) < frameHeaderSize {
		return nil, fmt.Errorf("sample frame too short: %d", len(frame))
	}
	if len(frame) > len(decoder.scratch) {
		return nil, fmt.Errorf("sample frame exceeds decoder capacity")
	}
	copy(decoder.scratch, frame)
	working := decoder.scratch[:len(frame)]
	channel := int(binary.BigEndian.Uint32(working[0:4]))
	sequence := binary.BigEndian.Uint64(working[4:12])
	voltage := math.Float64frombits(binary.BigEndian.Uint64(working[12:20]))
	temperature := math.Float64frombits(binary.BigEndian.Uint64(working[20:28]))
	payload := working[28:]
	return &model.Sample{
		ID:          uuid.NewString(),
		RunID:       runID,
		TrayID:      trayID,
		CellID:      cellID,
		Channel:     channel,
		Sequence:    sequence,
		Voltage:     voltage,
		Temperature: temperature,
		Payload:     payload,
		CapturedAt:  time.Now().UTC(),
	}, nil
}

func Encode(channel int, sequence uint64, voltage float64, temperature float64, payload []byte) []byte {
	frame := make([]byte, frameHeaderSize+len(payload))
	binary.BigEndian.PutUint32(frame[0:4], uint32(channel))
	binary.BigEndian.PutUint64(frame[4:12], sequence)
	binary.BigEndian.PutUint64(frame[12:20], math.Float64bits(voltage))
	binary.BigEndian.PutUint64(frame[20:28], math.Float64bits(temperature))
	copy(frame[28:], payload)
	return frame
}
