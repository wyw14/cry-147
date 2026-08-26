package cycler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-147/internal/cycler"
)

func TestFailedCyclerHandshakeReturnsAdmissionToken(t *testing.T) {
	admission := cycler.NewAdmission(1)
	failed := errors.New("offline device")
	if _, err := admission.Open(context.Background(), func(context.Context) error { return failed }); !errors.Is(err, failed) {
		t.Fatalf("unexpected handshake error %v", err)
	}
	if admission.Used() != 0 || admission.Active() != 0 {
		t.Fatalf("failed handshake retained capacity used=%d active=%d", admission.Used(), admission.Active())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	type result struct {
		lease *cycler.Lease
		err   error
	}
	resultChannel := make(chan result, 1)
	go func() {
		lease, err := admission.Open(ctx, func(context.Context) error { return nil })
		resultChannel <- result{lease: lease, err: err}
	}()
	var lease *cycler.Lease
	select {
	case value := <-resultChannel:
		if value.err != nil {
			t.Fatalf("healthy device could not acquire released token: %v", value.err)
		}
		lease = value.lease
	case <-time.After(200 * time.Millisecond):
		t.Fatal("healthy device remained blocked after failed handshake")
	}
	lease.Close()
	if admission.Used() != 0 {
		t.Fatalf("closed healthy session retained token")
	}
}
