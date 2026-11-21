package testkit

import "sync"

// Call is one recorded method call.
type Call struct {
	Method string
	Args   []any
}

// recorder records calls and returns errors scripted per method. The fakes
// embed it.
type recorder struct {
	mu    sync.Mutex
	calls []Call
	errs  map[string]error
}

// record stores the call and returns the error scripted for the method.
func (r *recorder) record(method string, args ...any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, Call{Method: method, Args: args})
	return r.errs[method]
}

// FailWith makes the method return err from now on (nil clears it).
func (r *recorder) FailWith(method string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.errs == nil {
		r.errs = map[string]error{}
	}
	r.errs[method] = err
}

// Calls returns a copy of all recorded calls in order.
func (r *recorder) Calls() []Call {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Call(nil), r.calls...)
}

// CallsTo returns the recorded calls of one method.
func (r *recorder) CallsTo(method string) []Call {
	var out []Call
	for _, c := range r.Calls() {
		if c.Method == method {
			out = append(out, c)
		}
	}
	return out
}
