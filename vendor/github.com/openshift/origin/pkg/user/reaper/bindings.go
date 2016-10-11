package reaper

import (
	"github.com/golang/glog"
	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"

	"github.com/openshift/origin/pkg/client"
)

// reapClusterBindings removes the subject from cluster-level role bindings
func reapClusterBindings(removedSubject kapi.ObjectReference, c client.ClusterRoleBindingsInterface) error {
	clusterBindings, err := c.ClusterRoleBindings().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	for _, binding := range clusterBindings.Items {
		retainedSubjects := []kapi.ObjectReference{}
		for _, subject := range binding.Subjects {
			if subject != removedSubject {
				retainedSubjects = append(retainedSubjects, subject)
			}
		}
		if len(retainedSubjects) != len(binding.Subjects) {
			updatedBinding := binding
			updatedBinding.Subjects = retainedSubjects
			if _, err := c.ClusterRoleBindings().Update(&updatedBinding); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update clusterrolebinding/%s: %v", binding.Name, err)
			}
		}
	}
	return nil
}

// reapNamespacedBindings removes the subject from namespaced role bindings
func reapNamespacedBindings(removedSubject kapi.ObjectReference, c client.RoleBindingsNamespacer) error {
	namespacedBindings, err := c.RoleBindings(kapi.NamespaceAll).List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	for _, binding := range namespacedBindings.Items {
		retainedSubjects := []kapi.ObjectReference{}
		for _, subject := range binding.Subjects {
			if subject != removedSubject {
				retainedSubjects = append(retainedSubjects, subject)
			}
		}
		if len(retainedSubjects) != len(binding.Subjects) {
			updatedBinding := binding
			updatedBinding.Subjects = retainedSubjects
			if _, err := c.RoleBindings(binding.Namespace).Update(&updatedBinding); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update rolebinding/%s in %s: %v", binding.Name, binding.Namespace, err)
			}
		}
	}
	return nil
}
