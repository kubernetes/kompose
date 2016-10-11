package client

import (
	"fmt"

	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	unversioned_extensions "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/extensions/unversioned"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/openshift/origin/pkg/api/latest"
)

type delegatingScaleInterface struct {
	dcs    DeploymentConfigInterface
	scales kclient.ScaleInterface
}

type delegatingScaleNamespacer struct {
	dcNS    DeploymentConfigsNamespacer
	scaleNS kclient.ScaleNamespacer
}

func (c *delegatingScaleNamespacer) Scales(namespace string) unversioned_extensions.ScaleInterface {
	return &delegatingScaleInterface{
		dcs:    c.dcNS.DeploymentConfigs(namespace),
		scales: c.scaleNS.Scales(namespace),
	}
}

func NewDelegatingScaleNamespacer(dcNamespacer DeploymentConfigsNamespacer, sNamespacer kclient.ScaleNamespacer) unversioned_extensions.ScalesGetter {
	return &delegatingScaleNamespacer{
		dcNS:    dcNamespacer,
		scaleNS: sNamespacer,
	}
}

// Get takes the reference to scale subresource and returns the subresource or error, if one occurs.
func (c *delegatingScaleInterface) Get(kind string, name string) (result *extensions.Scale, err error) {
	switch {
	case kind == "DeploymentConfig":
		return c.dcs.GetScale(name)
	// TODO: This is borked because the interface for Get is broken. Kind is insufficient.
	case latest.IsKindInAnyOriginGroup(kind):
		return nil, errors.NewBadRequest(fmt.Sprintf("Kind %s has no Scale subresource", kind))
	default:
		return c.scales.Get(kind, name)
	}
}

// Update takes a scale subresource object, updates the stored version to match it, and
// returns the subresource or error, if one occurs.
func (c *delegatingScaleInterface) Update(kind string, scale *extensions.Scale) (result *extensions.Scale, err error) {
	switch {
	case kind == "DeploymentConfig":
		return c.dcs.UpdateScale(scale)
	// TODO: This is borked because the interface for Update is broken. Kind is insufficient.
	case latest.IsKindInAnyOriginGroup(kind):
		return nil, errors.NewBadRequest(fmt.Sprintf("Kind %s has no Scale subresource", kind))
	default:
		return c.scales.Update(kind, scale)
	}
}
