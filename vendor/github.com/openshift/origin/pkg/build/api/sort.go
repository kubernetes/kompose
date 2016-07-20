package api

// BuildSliceByCreationTimestamp implements sort.Interface for []Build
// based on the CreationTimestamp field.
type BuildSliceByCreationTimestamp []Build

func (b BuildSliceByCreationTimestamp) Len() int {
	return len(b)
}

func (b BuildSliceByCreationTimestamp) Less(i, j int) bool {
	return b[i].CreationTimestamp.Before(b[j].CreationTimestamp)
}

func (b BuildSliceByCreationTimestamp) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

// BuildPtrSliceByCreationTimestamp implements sort.Interface for []*Build
// based on the CreationTimestamp field.
type BuildPtrSliceByCreationTimestamp []*Build

func (b BuildPtrSliceByCreationTimestamp) Len() int {
	return len(b)
}

func (b BuildPtrSliceByCreationTimestamp) Less(i, j int) bool {
	return b[i].CreationTimestamp.Before(b[j].CreationTimestamp)
}

func (b BuildPtrSliceByCreationTimestamp) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
