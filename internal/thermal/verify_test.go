package thermal_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/wyw14/cry-147/internal/thermal"
)

type trackingBody struct {
	closed *atomic.Int32
}

func (body *trackingBody) Read(buffer []byte) (int, error) { return 0, io.EOF }
func (body *trackingBody) Close() error                    { body.closed.Add(1); return nil }

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestThermalClientClosesBodyForErrorResponses(t *testing.T) {
	closed := &atomic.Int32{}
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Status: "503 Service Unavailable", Header: make(http.Header), Body: &trackingBody{closed: closed}, Request: request}, nil
	})}
	client, err := thermal.NewClient("http://thermal.local", httpClient)
	if err != nil {
		t.Fatal(err)
	}
	const requests = 12
	errors := make(chan error, requests)
	var group sync.WaitGroup
	for index := 0; index < requests; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, err := client.Reading(context.Background(), "rest-a")
			errors <- err
		}()
	}
	group.Wait()
	close(errors)
	for err := range errors {
		if err == nil || !strings.Contains(err.Error(), "503") {
			t.Fatalf("expected status error, got %v", err)
		}
	}
	if closed.Load() != requests {
		t.Fatalf("closed %d of %d response bodies", closed.Load(), requests)
	}
}
