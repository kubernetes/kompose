package testutils

import (
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// CheckForHeadless is helper function for tests.
// It checks if all Services in objects are Headless Services and if there is at least one such Services.
func CheckForHeadless(objects []runtime.Object) error {
	serviceCreated := false
	for _, obj := range objects {
		if svc, ok := obj.(*v1.Service); ok {
			serviceCreated = true
			// Check if it is a headless services
			if svc.Spec.ClusterIP != "None" {
				return errors.New("this is not a Headless services")
			}
		}
	}
	if !serviceCreated {
		return errors.New("no Service created")
	}
	return nil
}

// CheckForHealthCheckLivenessAndReadiness check if has liveness and readiness in healthcheck configured.
func CheckForHealthCheckLivenessAndReadiness(objects []runtime.Object) error {
	serviceCreated := false
	for _, obj := range objects {
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			serviceCreated = true

			// Check if it is a headless services
			if deployment.Spec.Template.Spec.Containers[0].ReadinessProbe == nil {
				return errors.New("there is not a ReadinessProbe")
			}
			if deployment.Spec.Template.Spec.Containers[0].LivenessProbe == nil {
				return errors.New("there is not a LivenessGate")
			}
		}
	}
	if !serviceCreated {
		return errors.New("no Service created")
	}
	return nil
}
