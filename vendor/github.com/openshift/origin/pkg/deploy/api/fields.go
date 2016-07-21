package api

import "k8s.io/kubernetes/pkg/fields"

// DeploymentConfigToSelectableFields returns a label set that represents the object
func DeploymentConfigToSelectableFields(deploymentConfig *DeploymentConfig) fields.Set {
	return fields.Set{
		"metadata.name":      deploymentConfig.Name,
		"metadata.namespace": deploymentConfig.Namespace,
	}
}
