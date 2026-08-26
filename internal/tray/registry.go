package tray

import (
	"fmt"
	"sort"
	"sync"

	"github.com/wyw14/cry-147/internal/model"
)

type Listener func(model.Tray)

type Registry struct {
	mu        sync.RWMutex
	trays     map[string]*model.Tray
	listeners map[string]Listener
}

func NewRegistry() *Registry {
	return &Registry{trays: map[string]*model.Tray{}, listeners: map[string]Listener{}}
}

func (registry *Registry) Register(tray *model.Tray) error {
	if err := model.ValidateTray(tray); err != nil {
		return err
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if _, exists := registry.trays[tray.ID]; exists {
		return fmt.Errorf("tray %s already registered", tray.ID)
	}
	registry.trays[tray.ID] = tray.Clone()
	return nil
}

func (registry *Registry) Lookup(id string) (*model.Tray, bool) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	tray, ok := registry.trays[id]
	return tray.Clone(), ok
}

func (registry *Registry) Replace(tray *model.Tray) error {
	if err := model.ValidateTray(tray); err != nil {
		return err
	}
	registry.mu.Lock()
	if _, exists := registry.trays[tray.ID]; !exists {
		registry.mu.Unlock()
		return fmt.Errorf("tray %s not found", tray.ID)
	}
	registry.trays[tray.ID] = tray.Clone()
	snapshot := registry.trays[tray.ID].Clone()
	listeners := make([]Listener, 0, len(registry.listeners))
	for _, listener := range registry.listeners {
		listeners = append(listeners, listener)
	}
	for _, listener := range listeners {
		listener(*snapshot.Clone())
	}
	registry.mu.Unlock()
	return nil
}

func (registry *Registry) Subscribe(id string, listener Listener) func() {
	registry.mu.Lock()
	registry.listeners[id] = listener
	registry.mu.Unlock()
	return func() {
		registry.mu.Lock()
		delete(registry.listeners, id)
		registry.mu.Unlock()
	}
}

func (registry *Registry) List() []*model.Tray {
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	ids := make([]string, 0, len(registry.trays))
	for id := range registry.trays {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]*model.Tray, 0, len(ids))
	for _, id := range ids {
		out = append(out, registry.trays[id].Clone())
	}
	return out
}

func (registry *Registry) Restore(trays []*model.Tray) error {
	restored := make(map[string]*model.Tray, len(trays))
	for _, value := range trays {
		if err := model.ValidateTray(value); err != nil {
			return err
		}
		if _, exists := restored[value.ID]; exists {
			return fmt.Errorf("duplicate recovered tray %s", value.ID)
		}
		restored[value.ID] = value.Clone()
	}
	registry.mu.Lock()
	registry.trays = restored
	registry.mu.Unlock()
	return nil
}
