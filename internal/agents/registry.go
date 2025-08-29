package agents

import (
	"errors"
	"sort"
	"sync"
)

// Registry provides threadsafe registration and selection of Agents by capability.
// It supports round-robin selection per task type to spread load across agents.
type Registry struct {
	mu     sync.RWMutex
	byName map[string]Agent   // agent name -> Agent
	byType map[string][]Agent // taskType -> agents that can handle it
	rrIdx  map[string]int     // taskType -> next round-robin index
}

// NewRegistry creates an empty agent registry.
func NewRegistry() *Registry {
	return &Registry{
		byName: make(map[string]Agent),
		byType: make(map[string][]Agent),
		rrIdx:  make(map[string]int),
	}
}

// Register adds an agent and indexes its capabilities. You can pass the task
// types explicitly, or leave empty and the registry will probe common types
// via CanHandle (useful if your agents can handle many or dynamic types).

// Prefer: Register(a, "grade", "notify", "recommend")
func (r *Registry) Register(a Agent, taskTypes ...string) error {
	if a == nil {
		return errors.New("nil agent")
	}
	name := a.Name()
	if name == "" {
		return errors.New("agent must have a non-empty Name()")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byName[name]; exists {
		return errors.New("agent already registered: " + name)
	}
	r.byName[name] = a

	// If explicit types not given, you can adapt this to your domain.
	// For now, only index provided types to avoid guessing.
	for _, t := range dedupe(taskTypes) {
		if !a.CanHandle(t) {
			// Skip silently; you may choose to error instead.
			continue
		}
		r.byType[t] = append(r.byType[t], a)
	}

	// Keep stable order for deterministic RR across runs (optional).
	for t := range r.byType {
		sort.SliceStable(r.byType[t], func(i, j int) bool {
			return r.byType[t][i].Name() < r.byType[t][j].Name()
		})
	}
	return nil
}

// Deregister removes an agent by name and unindexes it from all task types.
func (r *Registry) Deregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.byName[name]
	if !ok {
		return false
	}
	delete(r.byName, name)

	// Remove from all type lists
	for t, list := range r.byType {
		newList := list[:0]
		for _, ag := range list {
			if ag.Name() != name {
				newList = append(newList, ag)
			}
		}
		if len(newList) == 0 {
			delete(r.byType, t)
			delete(r.rrIdx, t)
		} else {
			r.byType[t] = newList
			// Clamp RR index
			if r.rrIdx[t] >= len(newList) {
				r.rrIdx[t] = 0
			}
		}
	}
	return true
}

// Get returns an agent by its unique name.
func (r *Registry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.byName[name]
	return a, ok
}

// Select chooses an agent that can handle the given task type.
// Uses round-robin across the set to balance load.
func (r *Registry) Select(taskType string) (Agent, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	list := r.byType[taskType]
	if len(list) == 0 {
		// Fallback: search all agents that say they can handle it (in case not indexed).
		for _, a := range r.byName {
			if a.CanHandle(taskType) {
				list = append(list, a)
			}
		}
		if len(list) == 0 {
			return nil, false
		}
		r.byType[taskType] = list
	}

	i := r.rrIdx[taskType] % len(list)
	a := list[i]
	r.rrIdx[taskType] = (i + 1) % len(list)
	return a, true
}

// List returns a snapshot of all registered agents.
func (r *Registry) List() []Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Agent, 0, len(r.byName))
	for _, a := range r.byName {
		out = append(out, a)
	}
	// Optional: sort by name for stable output
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
