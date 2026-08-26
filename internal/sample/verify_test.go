package sample_test

import (
	"bytes"
	"testing"

	"github.com/wyw14/cry-147/internal/model"
	"github.com/wyw14/cry-147/internal/sample"
)

func TestPooledSamplesRemainStableForAsyncConsumers(t *testing.T) {
	bus := sample.NewCoordinator()
	received := make(chan *model.Sample, 2)
	bus.Subscribe("slow-quality", func(value *model.Sample) error {
		received <- value
		return nil
	})
	input := model.Sample{ID: "sample-11", RunID: "run-11", TrayID: "tray-11", CellID: "cell-11", Channel: 11, Sequence: 42, Voltage: 3.65, Temperature: 26.5, Payload: []byte{4, 2, 1}}
	if err := bus.PublishPooled(input); err != nil {
		t.Fatal(err)
	}
	value := <-received
	if value.Sequence != 42 || value.Channel != 11 || value.Temperature != 26.5 || !bytes.Equal(value.Payload, []byte{4, 2, 1}) {
		t.Fatalf("consumer observed recycled sample %+v", value)
	}
	if err := bus.PublishPooled(model.Sample{ID: "sample-next", RunID: "run-11", Channel: 12, Sequence: 43, Payload: []byte{9}}); err != nil {
		t.Fatal(err)
	}
	if value.Sequence != 42 || !bytes.Equal(value.Payload, []byte{4, 2, 1}) {
		t.Fatalf("later pool reuse mutated retained sample %+v", value)
	}
}
