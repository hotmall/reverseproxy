package reverseproxy

import (
	"net/http"
	"sort"
	"strings"
	"sync"
)

type server struct {
	Title   string           `yaml:"title"`
	Version string           `yaml:"version"`
	BaseUri string           `yaml:"baseUri"`
	Target  string           `yaml:"target"`
	Proxy   map[string]proxy `yaml:"proxy"`
}

func (s *server) reset() {
	s.Title = ""
	s.Version = ""
	s.BaseUri = ""
	s.Target = ""
	s.Proxy = nil
}

type proxy struct {
	Methods []string `yaml:"methods,flow"`
	Pass    string   `yaml:"pass"`
}

type muxEntry struct {
	pattern string
	proxy   proxy
	h       http.Handler
}

func appendSorted(es []muxEntry, e muxEntry) []muxEntry {
	n := len(es)
	i := sort.Search(n, func(i int) bool {
		return len(es[i].pattern) < len(e.pattern)
	})
	if i == n {
		return append(es, e)
	}
	// we now know that i points at where we want to insert
	es = append(es, muxEntry{}) // try to grow the slice in place, any entry works.
	copy(es[i+1:], es[i:])      // Move shorter entries down
	es[i] = e
	return es
}

// proxyMux is an proxy multiplexer.
type proxyMux struct {
	mu sync.RWMutex
	m  map[string]muxEntry
	es []muxEntry // slice of entries sorted from longest to shortest.
}

var defaultProxyMux proxyMux

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *proxyMux) handle(pattern string, proxy proxy, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("proxy: invalid pattern")
	}
	if handler == nil {
		panic("proxy: nil handler")
	}
	if _, exist := mux.m[pattern]; exist {
		panic("proxy: multiple registrations for " + pattern)
	}

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}
	e := muxEntry{h: handler, proxy: proxy, pattern: pattern}
	mux.m[pattern] = e
	if pattern[len(pattern)-1] == '/' {
		mux.es = appendSorted(mux.es, e)
	}
}

func (mux *proxyMux) match(path string) (pattern string, proxy proxy, handler http.Handler) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Check for exact match first.
	v, ok := mux.m[path]
	if ok {
		pattern = v.pattern
		proxy = v.proxy
		handler = v.h
		return
	}

	// Check for longest valid match.  mux.es contains all patterns
	// that end in / sorted from longest to shortest.
	for _, e := range mux.es {
		if strings.HasPrefix(path, e.pattern) {
			return e.pattern, e.proxy, e.h
		}
	}
	return
}
