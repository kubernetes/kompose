/*
Copyright 2016 Skippbox, Ltd All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openshift

import (
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/transformer/kubernetes"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

type OpenShift struct {
}

// initDeploymentConfig initialize OpenShifts DeploymentConfig object
func initDeploymentConfig(name string, service kobject.ServiceConfig, replicas int) *deployapi.DeploymentConfig {
	dc := &deployapi.DeploymentConfig{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"service": name},
		},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: int32(replicas),
			Selector: map[string]string{"service": name},
			//UniqueLabelKey: p.Name,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{"service": name},
				},
				Spec: api.PodSpec{
					Containers: []api.Container{
						{
							Name:  name,
							Image: service.Image,
						},
					},
				},
			},
		},
	}
	return dc
}

// Transform maps komposeObject to openshift objects
// returns objects that are already sorted in the way that Services are first
func (k *OpenShift) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) []runtime.Object {
	// this will hold all the converted data
	var allobjects []runtime.Object

	for name, service := range komposeObject.ServiceConfigs {
		objects := kubernetes.CreateKubernetesObjects(name, service, opt)

		if opt.CreateDeploymentConfig {
			objects = append(objects, initDeploymentConfig(name, service, opt.Replicas)) // OpenShift DeploymentConfigs
		}

		// If ports not provided in configuration we will not make service
		if kubernetes.PortsExist(name, service) {
			svc := kubernetes.CreateService(name, service, objects)
			objects = append(objects, svc)
		}

		kubernetes.UpdateKubernetesObjects(name, service, objects)

		allobjects = append(allobjects, objects...)
	}
	// sort all object so Services are first
	kubernetes.SortServicesFirst(&allobjects)
	return allobjects
}
