package docker

import (
	"encoding/json"

	"github.com/docker/libcompose/utils"
)

// Label represents a docker label.
type Label string

// Libcompose default labels.
const (
	NAME    = Label("io.docker.compose.name")
	PROJECT = Label("io.docker.compose.project")
	SERVICE = Label("io.docker.compose.service")
	HASH    = Label("io.docker.compose.config-hash")
)

// EqString returns a label json string representation with the specified value.
func (f Label) EqString(value string) string {
	return utils.LabelFilterString(string(f), value)
}

// Eq returns a label map representation with the specified value.
func (f Label) Eq(value string) map[string][]string {
	return utils.LabelFilter(string(f), value)
}

// AndString returns a json list of labels by merging the two specified values (left and right) serialized as string.
func AndString(left, right string) string {
	leftMap := map[string][]string{}
	rightMap := map[string][]string{}

	// Ignore errors
	json.Unmarshal([]byte(left), &leftMap)
	json.Unmarshal([]byte(right), &rightMap)

	for k, v := range rightMap {
		existing, ok := leftMap[k]
		if ok {
			leftMap[k] = append(existing, v...)
		} else {
			leftMap[k] = v
		}
	}

	result, _ := json.Marshal(leftMap)

	return string(result)
}

// And returns a map of labels by merging the two specified values (left and right).
func And(left, right map[string][]string) map[string][]string {
	result := map[string][]string{}
	for k, v := range left {
		result[k] = v
	}

	for k, v := range right {
		existing, ok := result[k]
		if ok {
			result[k] = append(existing, v...)
		} else {
			result[k] = v
		}
	}

	return result
}

// Str returns the label name.
func (f Label) Str() string {
	return string(f)
}
