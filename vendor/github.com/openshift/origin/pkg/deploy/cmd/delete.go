package cmd

import (
	"time"

	"github.com/golang/glog"
	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
	kutil "k8s.io/kubernetes/pkg/util"
	"k8s.io/kubernetes/pkg/util/wait"

	"github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/openshift/origin/pkg/deploy/util"
)

// NewDeploymentConfigReaper returns a new reaper for deploymentConfigs
func NewDeploymentConfigReaper(oc client.Interface, kc kclient.Interface) kubectl.Reaper {
	return &DeploymentConfigReaper{oc: oc, kc: kc, pollInterval: kubectl.Interval, timeout: kubectl.Timeout}
}

// DeploymentConfigReaper implements the Reaper interface for deploymentConfigs
type DeploymentConfigReaper struct {
	oc                    client.Interface
	kc                    kclient.Interface
	pollInterval, timeout time.Duration
}

// pause marks the deployment configuration as paused to avoid triggering new
// deployments.
func (reaper *DeploymentConfigReaper) pause(namespace, name string) (*deployapi.DeploymentConfig, error) {
	return client.UpdateConfigWithRetries(reaper.oc, namespace, name, func(d *deployapi.DeploymentConfig) {
		d.Spec.RevisionHistoryLimit = kutil.Int32Ptr(0)
		d.Spec.Replicas = 0
		d.Spec.Paused = true
	})
}

// Stop scales a replication controller via its deployment configuration down to
// zero replicas, waits for all of them to get deleted and then deletes both the
// replication controller and its deployment configuration.
func (reaper *DeploymentConfigReaper) Stop(namespace, name string, timeout time.Duration, gracePeriod *kapi.DeleteOptions) error {
	// Pause the deployment configuration to prevent the new deployments from
	// being triggered.
	config, err := reaper.pause(namespace, name)
	configNotFound := kerrors.IsNotFound(err)
	if err != nil && !configNotFound {
		return err
	}

	var (
		isPaused bool
		legacy   bool
	)
	// Determine if the deployment config controller noticed the pause.
	if !configNotFound {
		if err := wait.Poll(1*time.Second, 1*time.Minute, func() (bool, error) {
			dc, err := reaper.oc.DeploymentConfigs(namespace).Get(name)
			if err != nil {
				return false, err
			}
			isPaused = dc.Spec.Paused
			return dc.Status.ObservedGeneration >= config.Generation, nil
		}); err != nil {
			return err
		}

		// If we failed to pause the deployment config, it means we are talking to
		// old API that does not support pausing. In that case, we delete the
		// deployment config to stay backward compatible.
		if !isPaused {
			if err := reaper.oc.DeploymentConfigs(namespace).Delete(name); err != nil {
				return err
			}
			// Setting this to true avoid deleting the config at the end.
			legacy = true
		}
	}

	// Clean up deployments related to the config. Even if the deployment
	// configuration has been deleted, we want to sweep the existing replication
	// controllers and clean them up.
	options := kapi.ListOptions{LabelSelector: util.ConfigSelector(name)}
	rcList, err := reaper.kc.ReplicationControllers(namespace).List(options)
	if err != nil {
		return err
	}
	rcReaper, err := kubectl.ReaperFor(kapi.Kind("ReplicationController"), reaper.kc)
	if err != nil {
		return err
	}

	// If there is neither a config nor any deployments, nor any deployer pods, we can return NotFound.
	deployments := rcList.Items

	if configNotFound && len(deployments) == 0 {
		return kerrors.NewNotFound(kapi.Resource("deploymentconfig"), name)
	}

	for _, rc := range deployments {
		if err = rcReaper.Stop(rc.Namespace, rc.Name, timeout, gracePeriod); err != nil {
			// Better not error out here...
			glog.Infof("Cannot delete ReplicationController %s/%s for deployment config %s/%s: %v", rc.Namespace, rc.Name, namespace, name, err)
		}

		// Only remove deployer pods when the deployment was failed. For completed
		// deployment the pods should be already deleted.
		if !util.IsFailedDeployment(&rc) {
			continue
		}

		// Delete all deployer and hook pods
		options = kapi.ListOptions{LabelSelector: util.DeployerPodSelector(rc.Name)}
		podList, err := reaper.kc.Pods(rc.Namespace).List(options)
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			err := reaper.kc.Pods(pod.Namespace).Delete(pod.Name, gracePeriod)
			if err != nil {
				// Better not error out here...
				glog.Infof("Cannot delete lifecycle Pod %s/%s for deployment config %s/%s: %v", pod.Namespace, pod.Name, namespace, name, err)
			}
		}
	}

	// Nothing to delete or we already deleted the deployment config because we
	// failed to pause.
	if configNotFound || legacy {
		return nil
	}

	return reaper.oc.DeploymentConfigs(namespace).Delete(name)
}
