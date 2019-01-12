package mbd

import (
	"strings"
)

// singleGet is a generic getter for single-value maps.
type singleGet struct {
	original  map[string]string
	lowercase map[string]string
}

// NewSingleGet initializes a new singleGet.
func newSingleGet(original map[string]string) *singleGet {
	lowercase := make(map[string]string, len(original))
	for k, v := range original {
		lowercase[strings.ToLower(k)] = v
	}
	return &singleGet{
		original:  original,
		lowercase: lowercase,
	}
}

// Map returns the original values as a map.
func (s *singleGet) Map() map[string]string {
	return s.original
}

// Get returns the value corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns "".
func (s *singleGet) Get(k string) string {
	return s.lowercase[strings.ToLower(k)]
}

// multiGet is a generic getter for multi-value maps.
type multiGet struct {
	original       map[string]string
	originalMulti  map[string][]string
	lowercaseMulti map[string][]string
}

// newMultiGet initializes a new multiGet.
func newMultiGet(original map[string]string, originalMulti map[string][]string) *multiGet {
	// fix for SAM local: it doesn't populate multimaps
	if len(originalMulti) < len(original) {
		originalMulti = map[string][]string{}
		for k, v := range original {
			if v == "" {
				originalMulti[k] = []string{}
			} else {
				originalMulti[k] = []string{v}
			}
		}
	}

	lowercaseMulti := make(map[string][]string, len(originalMulti))
	for k, v := range originalMulti {
		lowercaseMulti[strings.ToLower(k)] = v
	}

	return &multiGet{
		original:       original,
		originalMulti:  originalMulti,
		lowercaseMulti: lowercaseMulti,
	}
}

// Map returns the original values as single-value map, where the single value is the last occurring multi-value.
func (m *multiGet) Map() map[string]string {
	return m.original
}

// MapMulti returns the original values as a multi-value map.
func (m *multiGet) MapMulti() map[string][]string {
	return m.originalMulti
}

// Get returns a single value corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns "". If the key has multiple values, it returns the last one.
func (m *multiGet) Get(k string) string {
	v := m.GetMulti(k)
	if len(v) == 0 {
		return ""
	}
	return v[len(v)-1]
}

// Get returns the values corresponding to the given key, with case-insensitive matching.
// If the key is not present, it returns []string{}. If the key has multiple values, it returns all of them.
func (m *multiGet) GetMulti(k string) []string {
	v, ok := m.lowercaseMulti[strings.ToLower(k)]
	if !ok {
		return []string{}
	}
	return v
}
