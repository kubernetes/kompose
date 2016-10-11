package graph

import (
	"github.com/gonum/graph"
)

// Marker is a struct that describes something interesting on a Node
type Marker struct {
	// Node is the optional node that this message is attached to
	Node graph.Node
	// RelatedNodes is an optional list of other nodes that are involved in this marker.
	RelatedNodes []graph.Node

	// Severity indicates how important this problem is.
	Severity Severity
	// Key is a short string to identify this message
	Key string

	// Message is a human-readable string that describes what is interesting
	Message string
	// Suggestion is a human-readable string that holds advice for resolving this
	// marker.
	Suggestion Suggestion
}

// Severity indicates how important this problem is.
type Severity string

const (
	// InfoSeverity is interesting
	// TODO: Consider what to do with this once we revisit the graph api - currently not used.
	InfoSeverity Severity = "info"
	// WarningSeverity is probably wrong, but we aren't certain
	WarningSeverity Severity = "warning"
	// ErrorSeverity is definitely wrong, this won't work
	ErrorSeverity Severity = "error"
)

type Markers []Marker

// MarkerScanner is a function for analyzing a graph and finding interesting things in it
type MarkerScanner func(g Graph, f Namer) []Marker

func (m Markers) BySeverity(severity Severity) []Marker {
	ret := []Marker{}
	for i := range m {
		if m[i].Severity == severity {
			ret = append(ret, m[i])
		}
	}

	return ret
}

// FilterByNamespace returns all the markers that are not associated with missing nodes
// from other namespaces (other than the provided namespace).
func (m Markers) FilterByNamespace(namespace string) Markers {
	filtered := Markers{}

	for i := range m {
		markerNodes := []graph.Node{}
		markerNodes = append(markerNodes, m[i].Node)
		markerNodes = append(markerNodes, m[i].RelatedNodes...)
		hasCrossNamespaceLink := false
		for _, node := range markerNodes {
			if IsFromDifferentNamespace(namespace, node) {
				hasCrossNamespaceLink = true
				break
			}
		}
		if !hasCrossNamespaceLink {
			filtered = append(filtered, m[i])
		}
	}

	return filtered
}

type BySeverity []Marker

func (m BySeverity) Len() int      { return len(m) }
func (m BySeverity) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m BySeverity) Less(i, j int) bool {
	lhs := m[i]
	rhs := m[j]

	switch lhs.Severity {
	case ErrorSeverity:
		switch rhs.Severity {
		case ErrorSeverity:
			return false
		}
	case WarningSeverity:
		switch rhs.Severity {
		case ErrorSeverity, WarningSeverity:
			return false
		}
	case InfoSeverity:
		switch rhs.Severity {
		case ErrorSeverity, WarningSeverity, InfoSeverity:
			return false
		}
	}

	return true
}

type ByNodeID []Marker

func (m ByNodeID) Len() int      { return len(m) }
func (m ByNodeID) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ByNodeID) Less(i, j int) bool {
	if m[i].Node == nil {
		return true
	}
	if m[j].Node == nil {
		return false
	}
	return m[i].Node.ID() < m[j].Node.ID()
}

type ByKey []Marker

func (m ByKey) Len() int      { return len(m) }
func (m ByKey) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ByKey) Less(i, j int) bool {
	return m[i].Key < m[j].Key
}

type Suggestion string

func (s Suggestion) String() string {
	return string(s)
}
