package cycler

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var ErrTruncatedFrame = errors.New("truncated cycler frame")

type Stream struct {
	reader   *bufio.Reader
	maxFrame int
}

func NewStream(reader io.Reader, maxFrame int) *Stream {
	if maxFrame < 32 {
		maxFrame = 64 * 1024
	}
	return &Stream{reader: bufio.NewReader(reader), maxFrame: maxFrame}
}

func (stream *Stream) ReadFrame() ([]byte, error) {
	header := make([]byte, 4)
	// Use io.ReadFull instead of Read: a single Read may return fewer bytes
	// than requested when the underlying TCP stream delivers a frame split
	// across small packets. Treating a short read as a complete header/payload
	// desynchronizes the stream, producing bogus frame sizes and decoded
	// channel numbers that run into the tens of thousands. io.ReadFull keeps
	// reading until the buffer is full or EOF, so a frame can never be
	// reassembled from a partial read.
	if _, err := io.ReadFull(stream.reader, header); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("%w: header: %v", ErrTruncatedFrame, err)
	}
	size := int(binary.BigEndian.Uint32(header))
	if size <= 0 || size > stream.maxFrame {
		return nil, fmt.Errorf("frame size %d outside allowed range", size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(stream.reader, payload); err != nil {
		return nil, fmt.Errorf("%w: payload: %v", ErrTruncatedFrame, err)
	}
	return payload, nil
}

func EncodeFrame(payload []byte) []byte {
	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	copy(frame[4:], payload)
	return frame
}

func ReadAllFrames(reader io.Reader, maxFrame int) ([][]byte, error) {
	stream := NewStream(reader, maxFrame)
	frames := [][]byte{}
	for {
		frame, err := stream.ReadFrame()
		if errors.Is(err, io.EOF) {
			return frames, nil
		}
		if err != nil {
			return nil, err
		}
		frames = append(frames, frame)
	}
}
