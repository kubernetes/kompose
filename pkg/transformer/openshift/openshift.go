/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"strings"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer/kubernetes"

	"github.com/Sirupsen/logrus"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/runtime"

	oclient "github.com/openshift/origin/pkg/client"
	ocliconfig "github.com/openshift/origin/pkg/cmd/cli/config"

	"time"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploymentconfigreaper "github.com/openshift/origin/pkg/deploy/cmd"
	imageapi "github.com/openshift/origin/pkg/image/api"
	"k8s.io/kubernetes/pkg/kubectl"
)

type OpenShift struct {
	// Anonymous field allows for inheritance. We are basically inheriting
	// all of kubernetes.Kubernetes Methods and variables here. We'll overwite
	// some of those methods with our own for openshift.
	kubernetes.Kubernetes
}

// timeout is how long we'll wait for the termination of OpenShift resource to be successful
// used when undeploying resources from OpenShift
const TIMEOUT = 300

// list of all unsupported keys for this transformer
// Keys are names of variables in kobject struct.
// this is map to make searching for keys easier
// to make sure that unsupported key is not going to be reported twice
// by keeping record if already saw this key in another service
var unsupportedKey = map[string]bool{}

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
func (o *OpenShift) initImageStream(name string, service kobject.ServiceConfig) *imageapi.ImageStream {
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
func (o *OpenShift) initDeploymentConfig(name string, service kobject.ServiceConfig, replicas int) *deployapi.DeploymentConfig {
	tag := getImageTag(service.Image)
	containerName := []string{name}

	// Use ContainerName if it was set
	if service.ContainerName != "" {
		containerName = []string{service.ContainerName}
	}

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
						ContainerNames: containerName,
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
func (o *OpenShift) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) []runtime.Object {
	noSupKeys := o.Kubernetes.CheckUnsupportedKey(&komposeObject, unsupportedKey)
	for _, keyName := range noSupKeys {
		logrus.Warningf("OpenShift provider doesn't support %s key - ignoring", keyName)
	}
	// this will hold all the converted data
	var allobjects []runtime.Object

	for name, service := range komposeObject.ServiceConfigs {
		var objects []runtime.Object

		// Generate pod only and nothing more
		if service.Restart == "no" || service.Restart == "on-failure" {
			pod := o.InitPod(name, service)
			objects = append(objects, pod)
		} else {
			objects = o.CreateKubernetesObjects(name, service, opt)

			if opt.CreateDeploymentConfig {
				objects = append(objects, o.initDeploymentConfig(name, service, opt.Replicas)) // OpenShift DeploymentConfigs
				// create ImageStream after deployment (creating IS will trigger new deployment)
				objects = append(objects, o.initImageStream(name, service))
			}

			// If ports not provided in configuration we will not make service
			if o.PortsExist(name, service) {
				svc := o.CreateService(name, service, objects)
				objects = append(objects, svc)
			}
		}
		o.UpdateKubernetesObjects(name, service, &objects)

		allobjects = append(allobjects, objects...)
	}
	// If docker-compose has a volumes_from directive it will be handled here
	o.VolumesFrom(&allobjects, komposeObject)
	// sort all object so Services are first
	o.SortServicesFirst(&allobjects)
	return allobjects
}

// Create OpenShift client, returns OpenShift client
func (o *OpenShift) getOpenShiftClient() (*oclient.Client, error) {
	// initialize OpenShift Client
	loadingRules := ocliconfig.NewOpenShiftClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	oclientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, err
	}
	oclient := oclient.NewOrDie(oclientConfig)
	return oclient, nil
}

func (o *OpenShift) Deploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects := o.Transform(komposeObject, opt)
	pvcStr := " "
	if !opt.EmptyVols {
		pvcStr = " and PersistentVolumeClaims "
	}
	fmt.Println("We are going to create OpenShift DeploymentConfigs, Services" + pvcStr + "for your Dockerized application. \n" +
		"If you need different kind of resources, use the 'kompose convert' and 'oc create -f' commands instead. \n")

	oclient, err := o.getOpenShiftClient()
	if err != nil {
		return err
	}
	kclient, namespace, err := o.GetKubernetesClient()
	if err != nil {
		return err
	}

	for _, v := range objects {
		switch t := v.(type) {
		case *imageapi.ImageStream:
			_, err := oclient.ImageStreams(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created ImageStream: %s", t.Name)
		case *deployapi.DeploymentConfig:
			_, err := oclient.DeploymentConfigs(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created DeploymentConfig: %s", t.Name)
		case *api.Service:
			_, err := kclient.Services(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created Service: %s", t.Name)
		case *api.PersistentVolumeClaim:
			_, err := kclient.PersistentVolumeClaims(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created PersistentVolumeClaim: %s", t.Name)
		}
	}

	if !opt.EmptyVols {
		pvcStr = ",pvc"
	} else {
		pvcStr = ""
	}
	fmt.Println("\nYour application has been deployed to OpenShift. You can run 'oc get dc,svc,is" + pvcStr + "' for details.")

	return nil
}

func (o *OpenShift) Undeploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects := o.Transform(komposeObject, opt)

	oclient, err := o.getOpenShiftClient()
	if err != nil {
		return err
	}
	kclient, namespace, err := o.GetKubernetesClient()
	if err != nil {
		return err
	}

	for _, v := range objects {
		switch t := v.(type) {
		case *imageapi.ImageStream:
			//delete imageStream
			err = oclient.ImageStreams(namespace).Delete(t.Name)
			if err != nil {
				return err
			} else {
				logrus.Infof("Successfully deleted ImageStream: %s", t.Name)
			}
		case *deployapi.DeploymentConfig:
			// delete deploymentConfig
			dcreaper := deploymentconfigreaper.NewDeploymentConfigReaper(oclient, kclient)
			err := dcreaper.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
			if err != nil {
				return err
			} else {
				logrus.Infof("Successfully deleted DeploymentConfig: %s", t.Name)
			}
		case *api.Service:
			//delete svc
			rpService, err := kubectl.ReaperFor(api.Kind("Service"), kclient)
			if err != nil {
				return err
			}
			//FIXME: gracePeriod is nil
			err = rpService.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
			if err != nil {
				return err
			} else {
				logrus.Infof("Successfully deleted service: %s", t.Name)
			}
		case *api.PersistentVolumeClaim:
			// delete pvc
			err = kclient.PersistentVolumeClaims(namespace).Delete(t.Name)
			if err != nil {
				return err
			} else {
				logrus.Infof("Successfully deleted PersistentVolumeClaim: %s", t.Name)
			}
		}
	}
	return nil
}
