package util

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apimachinery/registered"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"
)

// ErrExit is a marker interface for cli commands indicating that the response has been processed
var ErrExit = fmt.Errorf("exit directly")

var commaSepVarsPattern = regexp.MustCompile(".*=.*,.*=.*")

// ReplaceCommandName recursively processes the examples in a given command to change a hardcoded
// command name (like 'kubectl' to the appropriate target name). It returns c.
func ReplaceCommandName(from, to string, c *cobra.Command) *cobra.Command {
	c.Example = strings.Replace(c.Example, from, to, -1)
	for _, sub := range c.Commands() {
		ReplaceCommandName(from, to, sub)
	}
	return c
}

// GetDisplayFilename returns the absolute path of the filename as long as there was no error, otherwise it returns the filename as-is
func GetDisplayFilename(filename string) string {
	if absName, err := filepath.Abs(filename); err == nil {
		return absName
	}

	return filename
}

// ResolveResource returns the resource type and name of the resourceString.
// If the resource string has no specified type, defaultResource will be returned.
func ResolveResource(defaultResource unversioned.GroupResource, resourceString string, mapper meta.RESTMapper) (unversioned.GroupResource, string, error) {
	if mapper == nil {
		return unversioned.GroupResource{}, "", errors.New("mapper cannot be nil")
	}

	var name string
	parts := strings.Split(resourceString, "/")
	switch len(parts) {
	case 1:
		name = parts[0]
	case 2:
		name = parts[1]

		// Allow specifying the group the same way kubectl does, as "resource.group.name"
		groupResource := unversioned.ParseGroupResource(parts[0])
		// normalize resource case
		groupResource.Resource = strings.ToLower(groupResource.Resource)

		gvr, err := mapper.ResourceFor(groupResource.WithVersion(""))
		if err != nil {
			return unversioned.GroupResource{}, "", err
		}
		return gvr.GroupResource(), name, nil
	default:
		return unversioned.GroupResource{}, "", fmt.Errorf("invalid resource format: %s", resourceString)
	}

	return defaultResource, name, nil
}

// convertItemsForDisplay returns a new list that contains parallel elements that have been converted to the most preferred external version
func convertItemsForDisplay(objs []runtime.Object, preferredVersions ...unversioned.GroupVersion) ([]runtime.Object, error) {
	ret := []runtime.Object{}

	for i := range objs {
		obj := objs[i]
		kind, _, err := kapi.Scheme.ObjectKind(obj)
		if err != nil {
			return nil, err
		}
		groupMeta, err := registered.Group(kind.Group)
		if err != nil {
			return nil, err
		}

		requestedVersion := unversioned.GroupVersion{}
		for _, preferredVersion := range preferredVersions {
			if preferredVersion.Group == kind.Group {
				requestedVersion = preferredVersion
				break
			}
		}

		actualOutputVersion := unversioned.GroupVersion{}
		for _, externalVersion := range groupMeta.GroupVersions {
			if externalVersion == requestedVersion {
				actualOutputVersion = externalVersion
				break
			}
			if actualOutputVersion.Empty() {
				actualOutputVersion = externalVersion
			}
		}

		convertedObject, err := kapi.Scheme.ConvertToVersion(obj, actualOutputVersion)
		if err != nil {
			return nil, err
		}

		ret = append(ret, convertedObject)
	}

	return ret, nil
}

// convertItemsForDisplayFromDefaultCommand returns a new list that contains parallel elements that have been converted to the most preferred external version
// TODO: move this function into the core factory PrintObjects method
// TODO: print-objects should have preferred output versions
func convertItemsForDisplayFromDefaultCommand(cmd *cobra.Command, objs []runtime.Object) ([]runtime.Object, error) {
	requested := kcmdutil.GetFlagString(cmd, "output-version")
	version, err := unversioned.ParseGroupVersion(requested)
	if err != nil {
		return nil, err
	}
	return convertItemsForDisplay(objs, version)
}

// VersionedPrintObject handles printing an object in the appropriate version by looking at 'output-version'
// on the command
func VersionedPrintObject(fn func(*cobra.Command, meta.RESTMapper, runtime.Object, io.Writer) error, c *cobra.Command, mapper meta.RESTMapper, out io.Writer) func(runtime.Object) error {
	return func(obj runtime.Object) error {
		// TODO: fold into the core printer functionality (preferred output version)
		if list, ok := obj.(*kapi.List); ok {
			var err error
			if list.Items, err = convertItemsForDisplayFromDefaultCommand(c, list.Items); err != nil {
				return err
			}
		} else {
			result, err := convertItemsForDisplayFromDefaultCommand(c, []runtime.Object{obj})
			if err != nil {
				return err
			}
			obj = result[0]
		}
		return fn(c, mapper, obj, out)
	}
}

func WarnAboutCommaSeparation(errout io.Writer, values []string, flag string) {
	for _, value := range values {
		if commaSepVarsPattern.MatchString(value) {
			fmt.Fprintf(errout, "warning: %s no longer accepts comma-separated lists of values. %q will be treated as a single key-value pair.\n", flag, value)
		}
	}
}
