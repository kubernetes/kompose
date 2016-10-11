package reaper

import (
	"time"

	"github.com/golang/glog"
	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"

	"github.com/openshift/origin/pkg/client"
)

func NewGroupReaper(
	groupClient client.GroupsInterface,
	clusterBindingClient client.ClusterRoleBindingsInterface,
	bindingClient client.RoleBindingsNamespacer,
	sccClient kclient.SecurityContextConstraintsInterface,
) kubectl.Reaper {
	return &GroupReaper{
		groupClient:          groupClient,
		clusterBindingClient: clusterBindingClient,
		bindingClient:        bindingClient,
		sccClient:            sccClient,
	}
}

type GroupReaper struct {
	groupClient          client.GroupsInterface
	clusterBindingClient client.ClusterRoleBindingsInterface
	bindingClient        client.RoleBindingsNamespacer
	sccClient            kclient.SecurityContextConstraintsInterface
}

// Stop on a reaper is actually used for deletion.  In this case, we'll delete referencing identities, clusterBindings, and bindings,
// then delete the group
func (r *GroupReaper) Stop(namespace, name string, timeout time.Duration, gracePeriod *kapi.DeleteOptions) error {
	removedSubject := kapi.ObjectReference{Kind: "Group", Name: name}

	if err := reapClusterBindings(removedSubject, r.clusterBindingClient); err != nil {
		return err
	}

	if err := reapNamespacedBindings(removedSubject, r.bindingClient); err != nil {
		return err
	}

	// Remove the group from sccs
	sccs, err := r.sccClient.SecurityContextConstraints().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	for _, scc := range sccs.Items {
		retainedGroups := []string{}
		for _, group := range scc.Groups {
			if group != name {
				retainedGroups = append(retainedGroups, group)
			}
		}
		if len(retainedGroups) != len(scc.Groups) {
			updatedSCC := scc
			updatedSCC.Groups = retainedGroups
			if _, err := r.sccClient.SecurityContextConstraints().Update(&updatedSCC); err != nil && !kerrors.IsNotFound(err) {
				glog.Infof("Cannot update scc/%s: %v", scc.Name, err)
			}
		}
	}

	// Remove the group
	if err := r.groupClient.Groups().Delete(name); err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	return nil
}
