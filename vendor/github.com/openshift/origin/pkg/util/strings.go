package util

import "sort"

// UniqueStrings returns a sorted, uniquified slice of the specified strings
func UniqueStrings(strings []string) []string {
	m := make(map[string]bool, len(strings))
	for _, s := range strings {
		m[s] = true
	}

	i := 0
	strings = make([]string, len(m), len(m))
	for s := range m {
		strings[i] = s
		i++
	}

	sort.Strings(strings)
	return strings
}
