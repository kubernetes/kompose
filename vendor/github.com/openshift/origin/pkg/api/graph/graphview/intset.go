package graphview

import (
	"sort"

	"k8s.io/kubernetes/pkg/util/sets"
)

type IntSet map[int]sets.Empty

// NewIntSet creates a IntSet from a list of values.
func NewIntSet(items ...int) IntSet {
	ss := IntSet{}
	ss.Insert(items...)
	return ss
}

// Insert adds items to the set.
func (s IntSet) Insert(items ...int) {
	for _, item := range items {
		s[item] = sets.Empty{}
	}
}

// Delete removes all items from the set.
func (s IntSet) Delete(items ...int) {
	for _, item := range items {
		delete(s, item)
	}
}

// Has returns true iff item is contained in the set.
func (s IntSet) Has(item int) bool {
	_, contained := s[item]
	return contained
}

// List returns the contents as a sorted string slice.
func (s IntSet) List() []int {
	res := make([]int, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	sort.IntSlice(res).Sort()
	return res
}
