package cycler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Handshake func(context.Context) error

type Admission struct {
	tokens chan struct{}
	mu     sync.Mutex
	active map[string]time.Time
}

type Lease struct {
	id        string
	admission *Admission
	once      sync.Once
}

func NewAdmission(limit int) *Admission {
	if limit < 1 {
		limit = 1
	}
	return &Admission{tokens: make(chan struct{}, limit), active: map[string]time.Time{}}
}

func (admission *Admission) Open(ctx context.Context, handshake Handshake) (*Lease, error) {
	select {
	case admission.tokens <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	lease := &Lease{id: uuid.NewString(), admission: admission}
	if err := handshake(ctx); err != nil {
		// The slot token was already acquired but the handshake failed, so the
		// caller will never receive a lease to Close. Release the token here to
		// avoid leaking admission slots on repeated offline-handshake failures,
		// which otherwise pin every slot in a "used" state with zero active
		// sessions and stall the tray in "opening".
		<-admission.tokens
		return nil, fmt.Errorf("cycler handshake: %w", err)
	}
	admission.mu.Lock()
	admission.active[lease.id] = time.Now().UTC()
	admission.mu.Unlock()
	return lease, nil
}

func (lease *Lease) Close() {
	if lease == nil || lease.admission == nil {
		return
	}
	lease.once.Do(func() {
		lease.admission.mu.Lock()
		delete(lease.admission.active, lease.id)
		lease.admission.mu.Unlock()
		<-lease.admission.tokens
	})
}

func (lease *Lease) ID() string {
	if lease == nil {
		return ""
	}
	return lease.id
}

func (admission *Admission) Used() int {
	return len(admission.tokens)
}

func (admission *Admission) Active() int {
	admission.mu.Lock()
	defer admission.mu.Unlock()
	return len(admission.active)
}
