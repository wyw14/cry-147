package cycler_test

import (
	"errors"
	"testing"

	"github.com/wyw14/cry-147/internal/cycler"
)

func TestChannelBusyIdentitySurvivesServiceBoundary(t *testing.T) {
	err := cycler.WrapAdapterError(8, cycler.ErrChannelBusy)
	if !errors.Is(err, cycler.ErrChannelBusy) {
		t.Fatalf("busy identity was lost across adapter boundary: %v", err)
	}
	if !cycler.Retryable(err) {
		t.Fatalf("campaign policy did not classify wrapped busy error as retryable")
	}
	permanent := errors.New("channel permanently failed")
	if cycler.Retryable(cycler.WrapAdapterError(8, permanent)) {
		t.Fatal("permanent error was misclassified as busy")
	}
}
