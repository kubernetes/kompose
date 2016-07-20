package extension

import (
	"fmt"

	"k8s.io/kubernetes/pkg/conversion"
	"k8s.io/kubernetes/pkg/runtime"
)

// Convert_runtime_Object_To_runtime_RawExtension attempts to convert runtime.Objects to the appropriate target, returning an error
// if there is insufficient information on the conversion scope to determine the target version.
func Convert_runtime_Object_To_runtime_RawExtension(c runtime.ObjectConvertor, in *runtime.Object, out *runtime.RawExtension, s conversion.Scope) error {
	if *in == nil {
		return nil
	}
	obj := *in

	switch obj.(type) {
	case *runtime.Unknown, *runtime.Unstructured:
		out.Raw = nil
		out.Object = obj
		return nil
	}

	switch t := s.Meta().Context.(type) {
	case runtime.GroupVersioner:
		converted, err := c.ConvertToVersion(obj, t)
		if err != nil {
			return err
		}
		out.Raw = nil
		out.Object = converted
	default:
		return fmt.Errorf("unrecognized conversion context for versioning: %#v", t)
	}
	return nil
}

// Convert_runtime_RawExtension_To_runtime_Object attempts to convert an incoming object into the
// appropriate output type.
func Convert_runtime_RawExtension_To_runtime_Object(c runtime.ObjectConvertor, in *runtime.RawExtension, out *runtime.Object, s conversion.Scope) error {
	if in == nil || in.Object == nil {
		return nil
	}

	switch in.Object.(type) {
	case *runtime.Unknown, *runtime.Unstructured:
		*out = in.Object
		return nil
	}

	switch t := s.Meta().Context.(type) {
	case runtime.GroupVersioner:
		converted, err := c.ConvertToVersion(in.Object, t)
		if err != nil {
			return err
		}
		in.Object = converted
		*out = converted
	default:
		return fmt.Errorf("unrecognized conversion context for conversion to internal: %#v (%T)", t, t)
	}
	return nil
}

// DecodeNestedRawExtensionOrUnknown
func DecodeNestedRawExtensionOrUnknown(d runtime.Decoder, ext *runtime.RawExtension) {
	if ext.Raw == nil || ext.Object != nil {
		return
	}
	obj, gvk, err := d.Decode(ext.Raw, nil, nil)
	if err != nil {
		unk := &runtime.Unknown{Raw: ext.Raw}
		if runtime.IsNotRegisteredError(err) {
			if _, gvk, err := d.Decode(ext.Raw, nil, unk); err == nil {
				unk.APIVersion = gvk.GroupVersion().String()
				unk.Kind = gvk.Kind
				ext.Object = unk
				return
			}
		}
		// TODO: record mime-type with the object
		if gvk != nil {
			unk.APIVersion = gvk.GroupVersion().String()
			unk.Kind = gvk.Kind
		}
		obj = unk
	}
	ext.Object = obj
}

// EncodeNestedRawExtension will encode the object in the RawExtension (if not nil) or
// return an error.
func EncodeNestedRawExtension(e runtime.Encoder, ext *runtime.RawExtension) error {
	if ext.Raw != nil || ext.Object == nil {
		return nil
	}
	data, err := runtime.Encode(e, ext.Object)
	if err != nil {
		return err
	}
	ext.Raw = data
	return nil
}
