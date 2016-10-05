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

package kubernetes

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/transformer"

	// install kubernetes api
	"k8s.io/kubernetes/pkg/api"
	_ "k8s.io/kubernetes/pkg/api/install"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
	//"k8s.io/kubernetes/pkg/controller/daemon"
	"time"

	"k8s.io/kubernetes/pkg/kubectl"
)

type Kubernetes struct {
}

// Init RC object
func InitRC(name string, service kobject.ServiceConfig, replicas int) *api.ReplicationController {
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: int32(replicas),
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: transformer.ConfigLabels(name),
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
	return rc
}

// Init Svc object
func InitSvc(name string, service kobject.ServiceConfig) *api.Service {
	svc := &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: api.ServiceSpec{
			Selector: transformer.ConfigLabels(name),
		},
	}
	return svc
}

// Init Deployment
func InitD(name string, service kobject.ServiceConfig, replicas int) *extensions.Deployment {
	dc := &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32(replicas),
			Template: api.PodTemplateSpec{
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

// Init DS object
func InitDS(name string, service kobject.ServiceConfig) *extensions.DaemonSet {
	ds := &extensions.DaemonSet{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: extensions.DaemonSetSpec{
			Template: api.PodTemplateSpec{
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
	return ds
}

// Initialize PersistentVolumeClaim
func CreatePVC(name string, mode string) *api.PersistentVolumeClaim {
	size, err := resource.ParseQuantity("100Mi")
	if err != nil {
		logrus.Fatalf("Error parsing size")
	}

	pvc := &api.PersistentVolumeClaim{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: api.PersistentVolumeClaimSpec{
			Resources: api.ResourceRequirements{
				Requests: api.ResourceList{
					api.ResourceStorage: size,
				},
			},
		},
	}

	if mode == "ro" {
		pvc.Spec.AccessModes = []api.PersistentVolumeAccessMode{"ReadWriteOnce"}
	} else {
		pvc.Spec.AccessModes = []api.PersistentVolumeAccessMode{"ReadWriteOnce"}
	}
	return pvc
}

// Configure the container ports.
func ConfigPorts(name string, service kobject.ServiceConfig) []api.ContainerPort {
	ports := []api.ContainerPort{}
	for _, port := range service.Port {
		ports = append(ports, api.ContainerPort{
			ContainerPort: port.ContainerPort,
			Protocol:      port.Protocol,
		})
	}

	return ports
}

// Configure the container service ports.
func ConfigServicePorts(name string, service kobject.ServiceConfig) []api.ServicePort {
	servicePorts := []api.ServicePort{}
	for _, port := range service.Port {
		if port.HostPort == 0 {
			port.HostPort = port.ContainerPort
		}
		var targetPort intstr.IntOrString
		targetPort.IntVal = port.ContainerPort
		targetPort.StrVal = strconv.Itoa(int(port.ContainerPort))
		servicePorts = append(servicePorts, api.ServicePort{
			Name:       strconv.Itoa(int(port.HostPort)),
			Protocol:   port.Protocol,
			Port:       port.HostPort,
			TargetPort: targetPort,
		})
	}
	return servicePorts
}

// Configure the container volumes.
func ConfigVolumes(name string, service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume, []*api.PersistentVolumeClaim) {
	volumesMount := []api.VolumeMount{}
	volumes := []api.Volume{}
	var pvc []*api.PersistentVolumeClaim

	var count int
	for _, volume := range service.Volumes {
		volumeName, host, container, mode, err := transformer.ParseVolume(volume)
		if err != nil {
			logrus.Warningf("Failed to configure container volume: %v", err)
			continue
		}
		if volumeName == "" {
			volumeName = fmt.Sprintf("%s-claim%d", name, count)
			count++
		}
		// check if ro/rw mode is defined, default rw
		readonly := len(mode) > 0 && mode == "ro"

		volmount := api.VolumeMount{
			Name:      volumeName,
			ReadOnly:  readonly,
			MountPath: container,
		}
		volumesMount = append(volumesMount, volmount)

		vol := api.Volume{
			Name: volumeName,
			VolumeSource: api.VolumeSource{
				PersistentVolumeClaim: &api.PersistentVolumeClaimVolumeSource{
					ClaimName: volumeName,
					ReadOnly:  readonly,
				},
			},
		}
		volumes = append(volumes, vol)

		if len(host) > 0 {
			logrus.Warningf("Volume mount on the host %q isn't supported - ignoring path on the host", host)
		}
		pvc = append(pvc, CreatePVC(volumeName, mode))
	}
	return volumesMount, volumes, pvc
}

// Configure the environment variables.
func ConfigEnvs(name string, service kobject.ServiceConfig) []api.EnvVar {
	envs := []api.EnvVar{}
	for _, v := range service.Environment {
		envs = append(envs, api.EnvVar{
			Name:  v.Name,
			Value: v.Value,
		})
	}

	return envs
}

// Generate a Kubernetes artifact for each input type service
func CreateKubernetesObjects(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions) []runtime.Object {
	var objects []runtime.Object

	if opt.CreateD {
		objects = append(objects, InitD(name, service, opt.Replicas))
	}
	if opt.CreateDS {
		objects = append(objects, InitDS(name, service))
	}
	if opt.CreateRC {
		objects = append(objects, InitRC(name, service, opt.Replicas))
	}

	return objects
}

// Transform maps komposeObject to k8s objects
// returns object that are already sorted in the way that Services are first
func (k *Kubernetes) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) []runtime.Object {
	// this will hold all the converted data
	var allobjects []runtime.Object

	for name, service := range komposeObject.ServiceConfigs {
		objects := CreateKubernetesObjects(name, service, opt)

		// If ports not provided in configuration we will not make service
		if PortsExist(name, service) {
			svc := CreateService(name, service, objects)
			objects = append(objects, svc)
		}

		UpdateKubernetesObjects(name, service, &objects)

		allobjects = append(allobjects, objects...)
	}
	// sort all object so Services are first
	SortServicesFirst(&allobjects)
	return allobjects
}

// Updates the given object with the given pod template update function and ObjectMeta update function
func UpdateController(obj runtime.Object, updateTemplate func(*api.PodTemplateSpec), updateMeta func(meta *api.ObjectMeta)) {
	switch t := obj.(type) {
	case *api.ReplicationController:
		if t.Spec.Template == nil {
			t.Spec.Template = &api.PodTemplateSpec{}
		}
		updateTemplate(t.Spec.Template)
		updateMeta(&t.ObjectMeta)
	case *extensions.Deployment:
		updateTemplate(&t.Spec.Template)
		updateMeta(&t.ObjectMeta)
	case *extensions.DaemonSet:
		updateTemplate(&t.Spec.Template)
		updateMeta(&t.ObjectMeta)
	case *deployapi.DeploymentConfig:
		updateTemplate(t.Spec.Template)
		updateMeta(&t.ObjectMeta)
	}
}

// Submit deployment and svc to k8s endpoint
func (k *Kubernetes) Deploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects := k.Transform(komposeObject, opt)

	fmt.Println("We are going to create Kubernetes deployments and services for your Dockerized application. \n" +
		"If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. \n")

	factory := cmdutil.NewFactory(nil)
	clientConfig, err := factory.ClientConfig()
	if err != nil {
		return err
	}
	namespace, _, err := factory.DefaultNamespace()
	if err != nil {
		return err
	}
	client := client.NewOrDie(clientConfig)

	for _, v := range objects {
		switch t := v.(type) {
		case *extensions.Deployment:
			_, err := client.Deployments(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created deployment: %s", t.Name)
		case *api.Service:
			_, err := client.Services(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created service: %s", t.Name)
		}
	}
	fmt.Println("\nYour application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods' for details.")

	return nil
}

func (k *Kubernetes) Undeploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {

	factory := cmdutil.NewFactory(nil)
	clientConfig, err := factory.ClientConfig()
	if err != nil {
		return err
	}
	namespace, _, err := factory.DefaultNamespace()
	if err != nil {
		return err
	}
	client := client.NewOrDie(clientConfig)

	// delete objects  from kubernetes
	for name := range komposeObject.ServiceConfigs {
		//delete svc
		rpService, err := kubectl.ReaperFor(api.Kind("Service"), client)
		if err != nil {
			return err
		}
		//FIXME: timeout = 300s, gracePeriod is nil
		err = rpService.Stop(namespace, name, 300*time.Second, nil)
		if err != nil {
			return err
		} else {
			logrus.Infof("Successfully deleted service: %s", name)
		}

		//delete deployment
		rpDeployment, err := kubectl.ReaperFor(extensions.Kind("Deployment"), client)
		if err != nil {
			return err
		}
		//FIXME: timeout = 300s, gracePeriod is nil
		err = rpDeployment.Stop(namespace, name, 300*time.Second, nil)
		if err != nil {
			return err
		} else {
			logrus.Infof("Successfully deleted deployment: %s", name)
		}
	}
	return nil
}
