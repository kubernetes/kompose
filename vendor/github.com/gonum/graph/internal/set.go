// Copyright ©2014 The gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"unsafe"

	"github.com/gonum/graph"
)

// IntSet is a set of integer identifiers.
type IntSet map[int]struct{}

// The simple accessor methods for Set are provided to allow ease of
// implementation change should the need arise.

// Add inserts an element into the set.
func (s IntSet) Add(e int) {
	s[e] = struct{}{}
}

// Has reports the existence of the element in the set.
func (s IntSet) Has(e int) bool {
	_, ok := s[e]
	return ok
}

// Remove deletes the specified element from the set.
func (s IntSet) Remove(e int) {
	delete(s, e)
}

// Count reports the number of elements stored in the set.
func (s IntSet) Count() int {
	return len(s)
}

// Same determines whether two sets are backed by the same store. In the
// current implementation using hash maps it makes use of the fact that
// hash maps (at least in the gc implementation) are passed as a pointer
// to a runtime Hmap struct.
//
// A map is not seen by the runtime as a pointer though, so we cannot
// directly compare the sets converted to unsafe.Pointer and need to take
// the sets' addressed and dereference them as pointers to some comparable
// type.
func Same(s1, s2 Set) bool {
	return *(*uintptr)(unsafe.Pointer(&s1)) == *(*uintptr)(unsafe.Pointer(&s2))
}

// A set is a set of nodes keyed in their integer identifiers.
type Set map[int]graph.Node

// The simple accessor methods for Set are provided to allow ease of
// implementation change should the need arise.

// Add inserts an element into the set.
func (s Set) Add(n graph.Node) {
	s[n.ID()] = n
}

// Remove deletes the specified element from the set.
func (s Set) Remove(e graph.Node) {
	delete(s, e.ID())
}

// Has reports the existence of the element in the set.
func (s Set) Has(n graph.Node) bool {
	_, ok := s[n.ID()]
	return ok
}

// Clear returns an empty set, possibly using the same backing store.
// Clear is not provided as a method since there is no way to replace
// the calling value if clearing is performed by a make(set). Clear
// should never be called without keeping the returned value.
func Clear(s Set) Set {
	if len(s) == 0 {
		return s
	}

	return make(Set)
}

// Copy performs a perfect copy from s1 to dst (meaning the sets will
// be equal).
func (dst Set) Copy(src Set) Set {
	if Same(src, dst) {
		return dst
	}

	if len(dst) > 0 {
		dst = make(Set, len(src))
	}

	for e, n := range src {
		dst[e] = n
	}

	return dst
}

// Equal reports set equality between the parameters. Sets are equal if
// and only if they have the same elements.
func Equal(s1, s2 Set) bool {
	if Same(s1, s2) {
		return true
	}

	if len(s1) != len(s2) {
		return false
	}

	for e := range s1 {
		if _, ok := s2[e]; !ok {
			return false
		}
	}

	return true
}

// Union takes the union of s1 and s2, and stores it in dst.
//
// The union of two sets, s1 and s2, is the set containing all the
// elements of each, for instance:
//
//     {a,b,c} UNION {d,e,f} = {a,b,c,d,e,f}
//
// Since sets may not have repetition, unions of two sets that overlap
// do not contain repeat elements, that is:
//
//     {a,b,c} UNION {b,c,d} = {a,b,c,d}
//
func (dst Set) Union(s1, s2 Set) Set {
	if Same(s1, s2) {
		return dst.Copy(s1)
	}

	if !Same(s1, dst) && !Same(s2, dst) {
		dst = Clear(dst)
	}

	if !Same(dst, s1) {
		for e, n := range s1 {
			dst[e] = n
		}
	}

	if !Same(dst, s2) {
		for e, n := range s2 {
			dst[e] = n
		}
	}

	return dst
}

// Intersect takes the intersection of s1 and s2, and stores it in dst.
//
// The intersection of two sets, s1 and s2, is the set containing all
// the elements shared between the two sets, for instance:
//
//     {a,b,c} INTERSECT {b,c,d} = {b,c}
//
// The intersection between a set and itself is itself, and thus
// effectively a copy operation:
//
//     {a,b,c} INTERSECT {a,b,c} = {a,b,c}
//
// The intersection between two sets that share no elements is the empty
// set:
//
//     {a,b,c} INTERSECT {d,e,f} = {}
//
func (dst Set) Intersect(s1, s2 Set) Set {
	var swap Set

	if Same(s1, s2) {
		return dst.Copy(s1)
	}
	if Same(s1, dst) {
		swap = s2
	} else if Same(s2, dst) {
		swap = s1
	} else {
		dst = Clear(dst)

		if len(s1) > len(s2) {
			s1, s2 = s2, s1
		}

		for e, n := range s1 {
			if _, ok := s2[e]; ok {
				dst[e] = n
			}
		}

		return dst
	}

	for e := range dst {
		if _, ok := swap[e]; !ok {
			delete(dst, e)
		}
	}

	return dst
}
