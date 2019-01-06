package getters

import (
	"strings"
)

// Single is a generic getter for single-value maps.
type Single struct {
	original  map[string]string
	lowercase map[string]string
}

// NewSingleGet initializes a new Single.
func NewSingle(original map[string]string) *Single {
	lowercase := make(map[string]string, len(original))
	for k, v := range original {
		lowercase[strings.ToLower(k)] = v
	}
	return &Single{
		original:  original,
		lowercase: lowercase,
	}
}

// Map returns the original values as a map.
func (s *Single) Map() map[string]string {
	return s.original
}

// Get returns the value corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns "".
func (s *Single) Get(k string) string {
	return s.lowercase[strings.ToLower(k)]
}

// Multi is a generic getter for multi-value maps.
type Multi struct {
	original       map[string]string
	originalMulti  map[string][]string
	lowercaseMulti map[string][]string
}

// NewMulti initializes a new Multi.
func NewMulti(original map[string]string, originalMulti map[string][]string) *Multi {
	lowercaseMulti := make(map[string][]string, len(originalMulti))
	for k, v := range originalMulti {
		lowercaseMulti[strings.ToLower(k)] = v
	}
	return &Multi{
		original:       original,
		originalMulti:  originalMulti,
		lowercaseMulti: lowercaseMulti,
	}
}

// Map returns the original values as single-value map, where the single value is the last occurring multi-value.
func (m *Multi) Map() map[string]string {
	return m.original
}

// MapMulti returns the original values as a multi-value map.
func (m *Multi) MapMulti() map[string][]string {
	return m.originalMulti
}

// Get returns a single value corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns "". If the key has multiple values, it returns the last one.
func (m *Multi) Get(k string) string {
	v := m.GetMulti(k)
	if len(v) == 0 {
		return ""
	}
	return v[len(v)-1]
}

// Get returns the values corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns []string{}. If the key has multiple values, it returns all of them.
func (m *Multi) GetMulti(k string) []string {
	v, ok := m.lowercaseMulti[strings.ToLower(k)]
	if !ok {
		return []string{}
	}
	return v
}
