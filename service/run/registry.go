package run

import "sync"

// Registry is the Runners of a master: the runners of the runs that are
// active, by run id. It is safe for concurrent use.
type Registry struct {
	mu      sync.Mutex
	runners map[uint]Runner
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{runners: make(map[uint]Runner)}
}

// Add remembers the runner of the run.
func (r *Registry) Add(runID uint, runner Runner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runners[runID] = runner
}

// Get returns the runner of the run, if it is active.
func (r *Registry) Get(runID uint) (Runner, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	runner, ok := r.runners[runID]
	return runner, ok
}

// Remove forgets the runner of the run.
func (r *Registry) Remove(runID uint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runners, runID)
}
