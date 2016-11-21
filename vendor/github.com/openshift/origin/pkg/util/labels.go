package util

import (
	"fmt"
	"reflect"

	kmeta "k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

// MergeInto flags
const (
	OverwriteExistingDstKey = 1 << iota
	ErrorOnExistingDstKey
	ErrorOnDifferentDstKeyValue
)

// AddObjectLabelsWithFlags will set labels on the target object.  Label overwrite behavior
// is controlled by the flags argument.
func AddObjectLabelsWithFlags(obj runtime.Object, labels labels.Set, flags int) error {
	if labels == nil {
		return nil
	}

	accessor, err := kmeta.Accessor(obj)

	if err != nil {
		if _, ok := obj.(*runtime.Unstructured); !ok {
			// error out if it's not possible to get an accessor and it's also not an unstructured object
			return err
		}
	} else {
		metaLabels := accessor.GetLabels()
		if metaLabels == nil {
			metaLabels = make(map[string]string)
		}

		switch objType := obj.(type) {
		case *deployapi.DeploymentConfig:
			if err := addDeploymentConfigNestedLabels(objType, labels, flags); err != nil {
				return fmt.Errorf("unable to add nested labels to %s/%s: %v", obj.GetObjectKind().GroupVersionKind(), accessor.GetName(), err)
			}
		}

		if err := MergeInto(metaLabels, labels, flags); err != nil {
			return fmt.Errorf("unable to add labels to %s/%s: %v", obj.GetObjectKind().GroupVersionKind(), accessor.GetName(), err)
		}

		accessor.SetLabels(metaLabels)

		return nil
	}

	// handle unstructured object
	// TODO: allow meta.Accessor to handle runtime.Unstructured
	if unstruct, ok := obj.(*runtime.Unstructured); ok && unstruct.Object != nil {
		// the presence of "metadata" is sufficient for us to apply the rules for Kube-like
		// objects.
		// TODO: add swagger detection to allow this to happen more effectively
		if obj, ok := unstruct.Object["metadata"]; ok {
			if m, ok := obj.(map[string]interface{}); ok {

				existing := make(map[string]string)
				if l, ok := m["labels"]; ok {
					if found, ok := interfaceToStringMap(l); ok {
						existing = found
					}
				}
				if err := MergeInto(existing, labels, flags); err != nil {
					return err
				}
				m["labels"] = mapToGeneric(existing)
			}
			return nil
		}

		// only attempt to set root labels if a root object called labels exists
		// TODO: add swagger detection to allow this to happen more effectively
		if obj, ok := unstruct.Object["labels"]; ok {
			existing := make(map[string]string)
			if found, ok := interfaceToStringMap(obj); ok {
				existing = found
			}
			if err := MergeInto(existing, labels, flags); err != nil {
				return err
			}
			unstruct.Object["labels"] = mapToGeneric(existing)
			return nil
		}
	}

	return nil

}

// AddObjectLabels adds new label(s) to a single runtime.Object, overwriting
// existing labels that have the same key.
func AddObjectLabels(obj runtime.Object, labels labels.Set) error {
	return AddObjectLabelsWithFlags(obj, labels, OverwriteExistingDstKey)
}

// AddObjectAnnotations adds new annotation(s) to a single runtime.Object
func AddObjectAnnotations(obj runtime.Object, annotations map[string]string) error {
	if len(annotations) == 0 {
		return nil
	}

	accessor, err := kmeta.Accessor(obj)

	if err != nil {
		if _, ok := obj.(*runtime.Unstructured); !ok {
			// error out if it's not possible to get an accessor and it's also not an unstructured object
			return err
		}
	} else {
		metaAnnotations := accessor.GetAnnotations()
		if metaAnnotations == nil {
			metaAnnotations = make(map[string]string)
		}

		switch objType := obj.(type) {
		case *deployapi.DeploymentConfig:
			if err := addDeploymentConfigNestedAnnotations(objType, annotations); err != nil {
				return fmt.Errorf("unable to add nested annotations to %s/%s: %v", obj.GetObjectKind().GroupVersionKind(), accessor.GetName(), err)
			}
		}

		MergeInto(metaAnnotations, annotations, OverwriteExistingDstKey)
		accessor.SetAnnotations(metaAnnotations)

		return nil
	}

	// handle unstructured object
	// TODO: allow meta.Accessor to handle runtime.Unstructured
	if unstruct, ok := obj.(*runtime.Unstructured); ok && unstruct.Object != nil {
		// the presence of "metadata" is sufficient for us to apply the rules for Kube-like
		// objects.
		// TODO: add swagger detection to allow this to happen more effectively
		if obj, ok := unstruct.Object["metadata"]; ok {
			if m, ok := obj.(map[string]interface{}); ok {

				existing := make(map[string]string)
				if l, ok := m["annotations"]; ok {
					if found, ok := interfaceToStringMap(l); ok {
						existing = found
					}
				}
				if err := MergeInto(existing, annotations, OverwriteExistingDstKey); err != nil {
					return err
				}
				m["annotations"] = mapToGeneric(existing)
			}
			return nil
		}

		// only attempt to set root annotations if a root object called annotations exists
		// TODO: add swagger detection to allow this to happen more effectively
		if obj, ok := unstruct.Object["annotations"]; ok {
			existing := make(map[string]string)
			if found, ok := interfaceToStringMap(obj); ok {
				existing = found
			}
			if err := MergeInto(existing, annotations, OverwriteExistingDstKey); err != nil {
				return err
			}
			unstruct.Object["annotations"] = mapToGeneric(existing)
			return nil
		}
	}

	return nil
}

// addDeploymentConfigNestedLabels adds new label(s) to a nested labels of a single DeploymentConfig object
func addDeploymentConfigNestedLabels(obj *deployapi.DeploymentConfig, labels labels.Set, flags int) error {
	if obj.Spec.Template.Labels == nil {
		obj.Spec.Template.Labels = make(map[string]string)
	}
	if err := MergeInto(obj.Spec.Template.Labels, labels, flags); err != nil {
		return fmt.Errorf("unable to add labels to Template.DeploymentConfig.Template.ControllerTemplate.Template: %v", err)
	}
	return nil
}

func addDeploymentConfigNestedAnnotations(obj *deployapi.DeploymentConfig, annotations map[string]string) error {
	if obj.Spec.Template == nil {
		return nil
	}

	if obj.Spec.Template.Annotations == nil {
		obj.Spec.Template.Annotations = make(map[string]string)
	}

	if err := MergeInto(obj.Spec.Template.Annotations, annotations, OverwriteExistingDstKey); err != nil {
		return fmt.Errorf("unable to add annotations to Template.DeploymentConfig.Template.ControllerTemplate.Template: %v", err)
	}
	return nil
}

// interfaceToStringMap extracts a map[string]string from a map[string]interface{}
func interfaceToStringMap(obj interface{}) (map[string]string, bool) {
	if obj == nil {
		return nil, false
	}
	lm, ok := obj.(map[string]interface{})
	if !ok {
		return nil, false
	}
	existing := make(map[string]string)
	for k, v := range lm {
		switch t := v.(type) {
		case string:
			existing[k] = t
		}
	}
	return existing, true
}

// mapToGeneric converts a map[string]string into a map[string]interface{}
func mapToGeneric(obj map[string]string) map[string]interface{} {
	if obj == nil {
		return nil
	}
	res := make(map[string]interface{})
	for k, v := range obj {
		res[k] = v
	}
	return res
}

// MergeInto merges items from a src map into a dst map.
// Returns an error when the maps are not of the same type.
// Flags:
// - ErrorOnExistingDstKey
//     When set: Return an error if any of the dst keys is already set.
// - ErrorOnDifferentDstKeyValue
//     When set: Return an error if any of the dst keys is already set
//               to a different value than src key.
// - OverwriteDstKey
//     When set: Overwrite existing dst key value with src key value.
func MergeInto(dst, src interface{}, flags int) error {
	dstVal := reflect.ValueOf(dst)
	srcVal := reflect.ValueOf(src)

	if dstVal.Kind() != reflect.Map {
		return fmt.Errorf("dst is not a valid map: %v", dstVal.Kind())
	}
	if srcVal.Kind() != reflect.Map {
		return fmt.Errorf("src is not a valid map: %v", srcVal.Kind())
	}
	if dstTyp, srcTyp := dstVal.Type(), srcVal.Type(); !dstTyp.AssignableTo(srcTyp) {
		return fmt.Errorf("type mismatch, can't assign '%v' to '%v'", srcTyp, dstTyp)
	}

	if dstVal.IsNil() {
		return fmt.Errorf("dst value is nil")
	}
	if srcVal.IsNil() {
		// Nothing to merge
		return nil
	}

	for _, k := range srcVal.MapKeys() {
		if dstVal.MapIndex(k).IsValid() {
			if flags&ErrorOnExistingDstKey != 0 {
				return fmt.Errorf("dst key already set (ErrorOnExistingDstKey=1), '%v'='%v'", k, dstVal.MapIndex(k))
			}
			if dstVal.MapIndex(k).String() != srcVal.MapIndex(k).String() {
				if flags&ErrorOnDifferentDstKeyValue != 0 {
					return fmt.Errorf("dst key already set to a different value (ErrorOnDifferentDstKeyValue=1), '%v'='%v'", k, dstVal.MapIndex(k))
				}
				if flags&OverwriteExistingDstKey != 0 {
					dstVal.SetMapIndex(k, srcVal.MapIndex(k))
				}
			}
		} else {
			dstVal.SetMapIndex(k, srcVal.MapIndex(k))
		}
	}

	return nil
}
