package reaper

import (
	"time"

	"github.com/golang/glog"
	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/kubectl"

	"github.com/openshift/origin/pkg/client"
)

func NewClusterRoleReaper(roleClient client.ClusterRolesInterface, clusterBindingClient client.ClusterRoleBindingsInterface, bindingClient client.RoleBindingsNamespacer) kubectl.Reaper {
	return &ClusterRoleReaper{
		roleClient:           roleClient,
		clusterBindingClient: clusterBindingClient,
		bindingClient:        bindingClient,
	}
}

type ClusterRoleReaper struct {
	roleClient           client.ClusterRolesInterface
	clusterBindingClient client.ClusterRoleBindingsInterface
	bindingClient        client.RoleBindingsNamespacer
}

// Stop on a reaper is actually used for deletion.  In this case, we'll delete referencing clusterroleclusterBindings
// then delete the clusterrole.
func (r *ClusterRoleReaper) Stop(namespace, name string, timeout time.Duration, gracePeriod *kapi.DeleteOptions) error {
	clusterBindings, err := r.clusterBindingClient.ClusterRoleBindings().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	for _, clusterBinding := range clusterBindings.Items {
		if clusterBinding.RoleRef.Name == name {
			if err := r.clusterBindingClient.ClusterRoleBindings().Delete(clusterBinding.Name); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot delete clusterrolebinding/%s: %v", clusterBinding.Name, err)
			}
		}
	}

	namespacedBindings, err := r.bindingClient.RoleBindings(kapi.NamespaceNone).List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	for _, namespacedBinding := range namespacedBindings.Items {
		if namespacedBinding.RoleRef.Namespace == kapi.NamespaceNone && namespacedBinding.RoleRef.Name == name {
			if err := r.bindingClient.RoleBindings(namespacedBinding.Namespace).Delete(namespacedBinding.Name); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot delete rolebinding/%s in %s: %v", namespacedBinding.Name, namespacedBinding.Namespace, err)
			}
		}
	}

	if err := r.roleClient.ClusterRoles().Delete(name); err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	return nil
}
