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
	"github.com/skippbox/kompose/pkg/transformer"
	"github.com/skippbox/kompose/pkg/transformer/kubernetes"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/sets"
)

type OpenShift struct {
}

// initDeploymentConfig initialize OpenShifts DeploymentConfig object
func initDeploymentConfig(services map[string]kobject.ServiceConfig, replicas int) *deployapi.DeploymentConfig {
	deploymentName := kubernetes.GetDeploymentName(services)
	dc := &deployapi.DeploymentConfig{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   deploymentName,
			Labels: transformer.ConfigLabels(deploymentName),
		},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: int32(replicas),
			Selector: transformer.ConfigLabels(deploymentName),
			//UniqueLabelKey: p.Name,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: transformer.ConfigLabels(deploymentName),
				},
				Spec: api.PodSpec{
					Containers: []api.Container{},
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
	// create a graph of dependecies
	d := make(map[string]sets.String)
	// this is a function that is specific to docker-compose directive `volumes_from`
	// later to add to dependency graph a user can create new function
	kubernetes.DependencyVolumesFrom(komposeObject, d)

	resolved := kubernetes.FindDependency(d)
	colocation := kubernetes.CalculateColocation(resolved)

	for _, svcnames := range colocation {
		svcConfigs := make(map[string]kobject.ServiceConfig)
		for _, svcname := range svcnames.List() {
			svcConfigs[svcname] = komposeObject.ServiceConfigs[svcname]
		}

		objects := kubernetes.CreateKubernetesObjects(svcConfigs, opt)

		if opt.CreateDeploymentConfig {
			objects = append(objects, initDeploymentConfig(svcConfigs, opt.Replicas)) // OpenShift DeploymentConfigs
		}

		// If ports not provided in configuration we will not make service
		svcs := kubernetes.InitSvc(svcConfigs)
		for _, svc := range svcs {
			objects = append(objects, svc)
		}
		kubernetes.UpdateKubernetesObjects(svcConfigs, objects, resolved)
		allobjects = append(allobjects, objects...)
	}
	// sort all object so Services are first
	kubernetes.SortServicesFirst(&allobjects)
	return allobjects
}
