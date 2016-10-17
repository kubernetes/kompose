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
	"errors"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	imageapi "github.com/openshift/origin/pkg/image/api"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer/kubernetes"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	"strings"
)

type OpenShift struct {
}

// getImageTag get tag name from image name
// if no tag is specified return 'latest'
func getImageTag(image string) string {
	p := strings.Split(image, ":")
	if len(p) == 2 {
		return p[1]
	} else {
		return "latest"
	}
}

// initImageStream initialize ImageStream object
func initImageStream(name string, service kobject.ServiceConfig) *imageapi.ImageStream {
	tag := getImageTag(service.Image)

	is := &imageapi.ImageStream{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: imageapi.ImageStreamSpec{
			Tags: map[string]imageapi.TagReference{
				tag: imageapi.TagReference{
					From: &api.ObjectReference{
						Kind: "DockerImage",
						Name: service.Image,
					},
				},
			},
		},
	}
	return is
}

// initDeploymentConfig initialize OpenShifts DeploymentConfig object
func initDeploymentConfig(name string, service kobject.ServiceConfig, replicas int) *deployapi.DeploymentConfig {
	tag := getImageTag(service.Image)

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
							Name: name,
							// Image will be set to ImageStream image by ImageChange trigger.
							Image: " ",
						},
					},
				},
			},
			Triggers: []deployapi.DeploymentTriggerPolicy{
				// Trigger new deploy when DeploymentConfig is created (config change)
				deployapi.DeploymentTriggerPolicy{
					Type: deployapi.DeploymentTriggerOnConfigChange,
				},
				deployapi.DeploymentTriggerPolicy{
					Type: deployapi.DeploymentTriggerOnImageChange,
					ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
						//Automatic - if new tag is detected - update image update inside the pod template
						Automatic:      true,
						ContainerNames: []string{name},
						From: api.ObjectReference{
							Name: name + ":" + tag,
							Kind: "ImageStreamTag",
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
			// create ImageStream after deployment (creating IS will trigger new deployment)
			objects = append(objects, initImageStream(name, service))
		}

		// If ports not provided in configuration we will not make service
		if kubernetes.PortsExist(name, service) {
			svc := kubernetes.CreateService(name, service, objects)
			objects = append(objects, svc)
		}

		kubernetes.UpdateKubernetesObjects(name, service, &objects)

		allobjects = append(allobjects, objects...)
	}
	// sort all object so Services are first
	kubernetes.SortServicesFirst(&allobjects)
	return allobjects
}

func (k *OpenShift) Deploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	return errors.New("Not Implemented")
}

func (k *OpenShift) Undeploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	return errors.New("Not Implemented")
}
