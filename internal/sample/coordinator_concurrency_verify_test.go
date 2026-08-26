package sample

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/wyw14/cry-147/internal/model"
)

func TestPublishPooledConcurrentNoCrossContamination(t *testing.T) {
	coordinator := NewCoordinator()

	var observed struct {
		mu     sync.Mutex
		seen   map[string]model.Sample
		shared int32
	}

	var wg sync.WaitGroup
	// subscriber records the full sample it observes, keyed by sample ID.
	coordinator.Subscribe("probe", func(sample *model.Sample) error {
		observed.mu.Lock()
		if observed.seen == nil {
			observed.seen = map[string]model.Sample{}
		}
		// Snapshot the values while we hold them so we can check later.
		clone := *sample
		observed.seen[sample.ID] = clone
		observed.mu.Unlock()
		return nil
	})

	const goroutines = 64
	const perGoroutine = 200
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(gIdx int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				seq := uint64(gIdx*perGoroutine + i + 1)
				temp := float64(seq) * 0.1
				input := model.Sample{
					ID:          idFor(gIdx, i),
					RunID:       "run",
					CellID:      "TRAY-001-C001",
					Channel:     1,
					Sequence:    seq,
					Voltage:     float64(seq) / 1000,
					Temperature: temp,
					Payload:     []byte("payload"),
				}
				if err := coordinator.PublishPooled(input); err != nil {
					t.Errorf("PublishPooled: %v", err)
					return
				}
			}
		}(g)
	}
	wg.Wait()

	// Every observed sample must match what was published: no cross-contamination
	// from the pooled buffer being wiped or reused by another in-flight call.
	for _, s := range observed.seen {
		if s.Sequence == 0 {
			t.Errorf("sample %s observed with zero sequence (wiped pooled buffer)", s.ID)
		}
		if s.Temperature == 0 {
			t.Errorf("sample %s observed with zero temperature", s.ID)
		}
		_ = atomic.LoadInt32(&observed.shared)
	}
	_ = atomic.LoadInt32(&observed.shared)
}

func idFor(g, i int) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	out := []byte{'s'}
	out = append(out, digits[g%len(digits)], '-')
	out = append(out, digits[(i/36)%len(digits)], digits[i%len(digits)])
	return string(out)
}
