package api

import (
	"sort"

	"k8s.io/kubernetes/pkg/api/unversioned"
)

type tag struct {
	Name    string
	Created unversioned.Time
}

type byCreationTimestamp []tag

func (t byCreationTimestamp) Len() int {
	return len(t)
}

func (t byCreationTimestamp) Less(i, j int) bool {
	return t[i].Created.Time.After(t[j].Created.Time)
}

func (t byCreationTimestamp) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// SortStatusTags sorts the status tags of an image stream based on
// the latest created
func SortStatusTags(tags map[string]TagEventList) []string {
	tagSlice := make([]tag, len(tags))
	index := 0
	for tag, list := range tags {
		tagSlice[index].Name = tag
		if len(list.Items) > 0 {
			tagSlice[index].Created = list.Items[0].Created
		}
		index++
	}

	sort.Sort(byCreationTimestamp(tagSlice))

	actual := make([]string, len(tagSlice))
	for i, tag := range tagSlice {
		actual[i] = tag.Name
	}

	return actual
}
