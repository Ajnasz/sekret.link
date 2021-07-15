package main

import (
	"net/http"
	"path"
	"sort"
	"sync"
)

func cleanPath(pattern string) string {
	if pattern == "" {
		return "/"
	}

	if pattern == "/" {
		return pattern
	}

	if pattern[0] != '/' {
		pattern = "/" + pattern
	}

	patternLen := len(pattern)

	if pattern[patternLen-1] == '/' {
		return pattern[0 : patternLen-1]
	}

	return pattern
}

func getPatternPartLen(pattern string) int {
	slashCount := 0

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '/' {
			slashCount++
		}
	}

	return slashCount
}

func appendSorted(es []MuxEntry, e MuxEntry) []MuxEntry {
	n := len(es)
	i := sort.Search(n, func(i int) bool {
		return getPatternPartLen(es[i].pattern) < getPatternPartLen(e.pattern)
	})

	if i == n {
		return append(es, e)
	}

	// we now know that i points at where we want to insert
	es = append(es, MuxEntry{}) // try to grow the slice in place, any entry works.
	copy(es[i+1:], es[i:])      // Move shorter entries down
	es[i] = e

	return es
}

func newMuxEntry(pattern string, handler http.Handler) MuxEntry {
	entry := MuxEntry{
		pattern: cleanPath(pattern),
		handler: handler,
	}
	entry.ValiatePattern()

	return entry
}

// MuxEntry A registered http request handler
type MuxEntry struct {
	handler http.Handler
	pattern string
	method  string
}

// ValiatePattern Validates that the given URL pattern is valid
func (m MuxEntry) ValiatePattern() {
	patternIter := pathIterator(m.pattern)

	for {
		value, ok := patternIter()

		if !ok {
			return
		}

		if value == "" {
			return
		}

		if isPathParameter(value) {
			if len(value) == 1 {
				panic("Invalid path parameter")
			}
		}
	}
}

// Method Set method for the entry
func (m *MuxEntry) Method(method string) *MuxEntry {
	m.method = method
	return m
}

// Post set http.Method to http.MethodPost
func (m *MuxEntry) Post() *MuxEntry {
	m.method = http.MethodPost
	return m
}

// IsMethodMatching Check if the given method is matching to the entry's method
func (m MuxEntry) IsMethodMatching(method string) bool {
	if m.method == "" {
		return true
	}

	return m.method == method
}

// Mux A struct handle http requests, match to URL patterns and execute
// handlers
type Mux struct {
	// entries
	es []MuxEntry
	m  map[string]MuxEntry
	mu *sync.RWMutex
}

// Handle Registers http handlers to a given URL pattern
func (mux *Mux) Handle(pattern string, handler http.Handler) *MuxEntry {
	if mux.mu == nil {
		mux.mu = &sync.RWMutex{}
	}
	mux.mu.Lock()
	defer mux.mu.Unlock()
	if pattern == "" {
		panic("mux: Invalid pattern")
	}

	if handler == nil {
		panic("mux: nil handler")
	}

	if mux.m == nil {
		mux.m = make(map[string]MuxEntry)
	}

	if _, exists := mux.m[pattern]; exists {
		panic("mux: multiple registrations for pattern: " + pattern)
	}

	entry := newMuxEntry(pattern, handler)
	mux.m[pattern] = entry
	mux.es = appendSorted(mux.es, entry)

	return &entry
}

func isEntryMatching(entry MuxEntry, r *http.Request) bool {
	if !entry.IsMethodMatching(r.Method) {
		return false
	}

	urlPath := r.URL.Path
	if urlPath[0] != '/' {
		urlPath = "/" + urlPath
	}
	return isPatternMatching(entry.pattern, urlPath)
}

func (mux Mux) getMatchingEntry(r *http.Request) *MuxEntry {
	for _, entry := range mux.es {
		if isEntryMatching(entry, r) {
			return &entry
		}
	}

	return nil
}

func (mux Mux) getMatchingHandler(r *http.Request) http.Handler {
	entry := mux.getMatchingEntry(r)
	if entry == nil {
		return nil
	}

	return entry.handler
}

func (mux Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := mux.getMatchingHandler(r)

	if handler == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r)
}

// GetVars returns variables defined to the URL pattern
// /foo/:var1/:var2
func (mux Mux) GetVars(r *http.Request) map[string]string {
	entry := mux.getMatchingEntry(r)

	if entry == nil {
		return nil
	}

	return parsePatternMatch(entry.pattern, r.URL.Path)
}

func parsePatternMatch(pattern string, str string) map[string]string {
	ret := make(map[string]string)
	patternIter := pathIterator(pattern)
	strIter := pathIterator(str)

	for {
		patternValue, patternOk := patternIter()

		if !patternOk {
			return ret
		}

		strValue, _ := strIter()

		if isPathParameter(patternValue) {
			key := getPathParameterKey(patternValue)
			ret[key] = strValue
		}
	}
}

func isPathParameter(str string) bool {
	return str[0] == ':'
}

func getPathParameterKey(str string) string {
	return str[1:]
}

func isPatternMatching(pattern string, str string) bool {
	if pattern == str {
		return true
	}
	patternIter := pathIterator(pattern)
	strIter := pathIterator(str)

	for {
		patternValue, patternOk := patternIter()
		strValue, strOk := strIter()

		if patternOk != strOk {
			return false
		}

		if !patternOk {
			return true
		}

		if isPathParameter(patternValue) {
			continue
		}

		if patternValue != strValue {
			return false
		}

	}
}

func pathIterator(str string) func() (string, bool) {
	return func() (string, bool) {
		dir, value := path.Split(str)

		if dir == "" && value == "" {
			return "", false
		}

		if dir != "" {
			str = dir[0 : len(dir)-1]
		}

		return value, true
	}
}
