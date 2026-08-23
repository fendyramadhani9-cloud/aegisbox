package runtime

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aegisbox/aegisbox/internal/executor"
)

var (
	ErrRuntimeNotFound      = errors.New("runtime not found for language")
	ErrRuntimeAlreadyExists = errors.New("runtime already registered")
)

// Runtime is an alias for the domain executor.RuntimeAdapter.
type Runtime = executor.RuntimeAdapter

// Registry manages registered language runtime adapters.
type Registry struct {
	mu       sync.RWMutex
	runtimes map[string]Runtime
}

var (
	defaultRegistry = NewRegistry()
)

// NewRegistry initializes an empty runtime registry.
func NewRegistry() *Registry {
	return &Registry{
		runtimes: make(map[string]Runtime),
	}
}

// DefaultRegistry returns the singleton registry populated with standard runtimes.
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// Register adds a runtime adapter to the registry.
func (r *Registry) Register(rt Runtime) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := rt.Name()
	if _, exists := r.runtimes[name]; exists {
		return fmt.Errorf("%w: %s", ErrRuntimeAlreadyExists, name)
	}
	r.runtimes[name] = rt
	return nil
}

// Get retrieves a runtime adapter by language name.
func (r *Registry) Get(name string) (Runtime, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rt, exists := r.runtimes[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRuntimeNotFound, name)
	}
	return rt, nil
}

// SupportedLanguages returns a list of all registered runtime names.
func (r *Registry) SupportedLanguages() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	langs := make([]string, 0, len(r.runtimes))
	for k := range r.runtimes {
		langs = append(langs, k)
	}
	return langs
}

func init() {
	_ = defaultRegistry.Register(NewPythonRuntime())
}
