package util

import (
	"path"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
)

// NoNamespaceKeyFunc is the default function for constructing etcd paths to a resource relative to prefix enforcing
// If a namespace is on context, it errors.
func NoNamespaceKeyFunc(ctx kapi.Context, prefix string, name string) (string, error) {
	ns, ok := kapi.NamespaceFrom(ctx)
	if ok && len(ns) > 0 {
		return "", kerrors.NewBadRequest("Namespace parameter is not allowed.")
	}
	if len(name) == 0 {
		return "", kerrors.NewBadRequest("Name parameter required.")
	}
	return path.Join(prefix, name), nil
}
