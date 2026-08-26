package cycler_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/wyw14/cry-147/internal/cycler"
)

type shortReader struct {
	reader *bytes.Reader
	limit  int
}

func (reader *shortReader) Read(buffer []byte) (int, error) {
	if len(buffer) > reader.limit {
		buffer = buffer[:reader.limit]
	}
	return reader.reader.Read(buffer)
}

func TestCyclerStreamReassemblesFragmentedFrames(t *testing.T) {
	first := []byte("first-complete-frame")
	second := []byte("second-complete-frame")
	wire := append(cycler.EncodeFrame(first), cycler.EncodeFrame(second)...)
	frames, err := cycler.ReadAllFrames(&shortReader{reader: bytes.NewReader(wire), limit: 2}, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 2 || !bytes.Equal(frames[0], first) || !bytes.Equal(frames[1], second) {
		t.Fatalf("fragmented stream decoded as %#v", frames)
	}
	truncated := cycler.NewStream(bytes.NewReader(cycler.EncodeFrame(first)[:7]), 1024)
	if _, err := truncated.ReadFrame(); err == nil || err == io.EOF {
		t.Fatalf("expected explicit truncated frame error, got %v", err)
	}
}
