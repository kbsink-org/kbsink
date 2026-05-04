package pluginreg

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kbsink-org/kbsink/pkg/core"
)

var (
	mu    sync.RWMutex
	names []string
	by    = map[string]core.Plugin{}
)

// Register adds a named plugin. The registry key is normalized from p.Name() (lower case, trimmed).
// Panics if the derived name is empty, or the same name is registered twice.
func Register(p core.Plugin) {
	if p == nil {
		panic("pluginreg: nil Plugin")
	}
	n := normalizeName(p.Name())
	if n == "" {
		panic("pluginreg: empty plugin name")
	}
	mu.Lock()
	defer mu.Unlock()
	if _, dup := by[n]; dup {
		panic(fmt.Sprintf("pluginreg: duplicate plugin %q", n))
	}
	by[n] = p
	names = append(names, n)
}

// Lookup returns the plugin for a name (case-insensitive).
func Lookup(name string) (core.Plugin, bool) {
	n := normalizeName(name)
	if n == "" {
		return nil, false
	}
	mu.RLock()
	defer mu.RUnlock()
	pl, ok := by[n]
	return pl, ok
}

// Names returns registered plugin names in registration order (for diagnostics).
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, len(names))
	copy(out, names)
	return out
}

func normalizeName(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}
