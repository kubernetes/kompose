package testutils

import (
	"errors"
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
