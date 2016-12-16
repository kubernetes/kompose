package api

import (
	"container/list"
	"reflect"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

// ClusterResourceQuota mirrors ResourceQuota at a cluster scope.  This object is easily convertible to
// synthetic ResourceQuota object to allow quota evaluation re-use.
type ClusterResourceQuota struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// Spec defines the desired quota
	Spec ClusterResourceQuotaSpec

	// Status defines the actual enforced quota and its current usage
	Status ClusterResourceQuotaStatus
}

// ClusterResourceQuotaSpec defines the desired quota restrictions
type ClusterResourceQuotaSpec struct {
	// Selector is the selector used to match projects.
	// It should only select active projects on the scale of dozens (though it can select
	// many more less active projects).  These projects will contend on object creation through
	// this resource.
	Selector ClusterResourceQuotaSelector

	// Quota defines the desired quota
	Quota kapi.ResourceQuotaSpec
}

// ClusterResourceQuotaSelector is used to select projects.  At least one of LabelSelector or AnnotationSelector
// must present.  If only one is present, it is the only selection criteria.  If both are specified,
// the project must match both restrictions.
type ClusterResourceQuotaSelector struct {
	// LabelSelector is used to select projects by label.
	LabelSelector *unversioned.LabelSelector

	// AnnotationSelector is used to select projects by annotation.
	AnnotationSelector map[string]string
}

// ClusterResourceQuotaStatus defines the actual enforced quota and its current usage
type ClusterResourceQuotaStatus struct {
	// Total defines the actual enforced quota and its current usage across all projects
	Total kapi.ResourceQuotaStatus

	// Namespaces slices the usage by project.  This division allows for quick resolution of
	// deletion reconcilation inside of a single project without requiring a recalculation
	// across all projects.  This map can be used to pull the deltas for a given project.
	Namespaces ResourceQuotasStatusByNamespace
}

// ClusterResourceQuotaList is a collection of ClusterResourceQuotas
type ClusterResourceQuotaList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of ClusterResourceQuotas
	Items []ClusterResourceQuota
}

// AppliedClusterResourceQuota mirrors ClusterResourceQuota at a project scope, for projection
// into a project.  It allows a project-admin to know which ClusterResourceQuotas are applied to
// his project and their associated usage.
type AppliedClusterResourceQuota struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	kapi.ObjectMeta

	// Spec defines the desired quota
	Spec ClusterResourceQuotaSpec

	// Status defines the actual enforced quota and its current usage
	Status ClusterResourceQuotaStatus
}

// AppliedClusterResourceQuotaList is a collection of AppliedClusterResourceQuotas
type AppliedClusterResourceQuotaList struct {
	unversioned.TypeMeta
	// Standard object's metadata.
	unversioned.ListMeta

	// Items is a list of AppliedClusterResourceQuota
	Items []AppliedClusterResourceQuota
}

// ResourceQuotasStatusByNamespace provides type correct methods
type ResourceQuotasStatusByNamespace struct {
	orderedMap orderedMap
}

func (o *ResourceQuotasStatusByNamespace) Insert(key string, value kapi.ResourceQuotaStatus) {
	o.orderedMap.Insert(key, value)
}

func (o *ResourceQuotasStatusByNamespace) Get(key string) (kapi.ResourceQuotaStatus, bool) {
	ret, ok := o.orderedMap.Get(key)
	if !ok {
		return kapi.ResourceQuotaStatus{}, ok
	}
	return ret.(kapi.ResourceQuotaStatus), ok
}

func (o *ResourceQuotasStatusByNamespace) Remove(key string) {
	o.orderedMap.Remove(key)
}

func (o *ResourceQuotasStatusByNamespace) OrderedKeys() *list.List {
	return o.orderedMap.OrderedKeys()
}

// DeepCopy implements a custom copy to correctly handle unexported fields
// Must match "func (t T) DeepCopy() T" for the deep copy generator to use it
func (o ResourceQuotasStatusByNamespace) DeepCopy() ResourceQuotasStatusByNamespace {
	out := ResourceQuotasStatusByNamespace{}
	for e := o.OrderedKeys().Front(); e != nil; e = e.Next() {
		namespace := e.Value.(string)
		instatus, _ := o.Get(namespace)
		if outstatus, err := kapi.Scheme.DeepCopy(instatus); err != nil {
			panic(err) // should never happen
		} else {
			out.Insert(namespace, outstatus.(kapi.ResourceQuotaStatus))
		}
	}
	return out
}

func init() {
	// Tell the reflection package how to compare our unexported type
	if err := kapi.Semantic.AddFuncs(
		func(o1, o2 ResourceQuotasStatusByNamespace) bool {
			return reflect.DeepEqual(o1.orderedMap, o2.orderedMap)
		},
		func(o1, o2 *ResourceQuotasStatusByNamespace) bool {
			if o1 == nil && o2 == nil {
				return true
			}
			if (o1 == nil) != (o2 == nil) {
				return false
			}
			return reflect.DeepEqual(o1.orderedMap, o2.orderedMap)
		},
	); err != nil {
		panic(err)
	}
}

// orderedMap is a very simple ordering a map tracking insertion order.  It allows fast and stable serializations
// for our encoding.  You could probably do something fancier with pointers to interfaces, but I didn't.
type orderedMap struct {
	backingMap  map[string]interface{}
	orderedKeys *list.List
}

// Insert puts something else in the map.  keys are ordered based on first insertion, not last touch.
func (o *orderedMap) Insert(key string, value interface{}) {
	if o.backingMap == nil {
		o.backingMap = map[string]interface{}{}
	}
	if o.orderedKeys == nil {
		o.orderedKeys = list.New()
	}

	if _, exists := o.backingMap[key]; !exists {
		o.orderedKeys.PushBack(key)
	}
	o.backingMap[key] = value
}

func (o *orderedMap) Get(key string) (interface{}, bool) {
	ret, ok := o.backingMap[key]
	return ret, ok
}

func (o *orderedMap) Remove(key string) {
	delete(o.backingMap, key)

	if o.orderedKeys == nil {
		return
	}
	for e := o.orderedKeys.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == key {
			o.orderedKeys.Remove(e)
			break
		}
	}
}

// OrderedKeys returns back the ordered keys.  This can be used to build a stable serialization
func (o *orderedMap) OrderedKeys() *list.List {
	if o.orderedKeys == nil {
		return list.New()
	}
	return o.orderedKeys
}
