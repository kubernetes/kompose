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
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer"
	deployapi "github.com/openshift/origin/pkg/deploy/api"

	// install kubernetes api
	"k8s.io/kubernetes/pkg/api"
	_ "k8s.io/kubernetes/pkg/api/install"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
	//"k8s.io/kubernetes/pkg/controller/daemon"
)

type Kubernetes struct {
	// the user provided options from the command line
	Opt kobject.ConvertOptions
}

// timeout is how long we'll wait for the termination of kubernetes resource to be successful
// used when undeploying resources from kubernetes
const TIMEOUT = 300

// Init RC object
func (k *Kubernetes) InitRC(name string, service kobject.ServiceConfig, replicas int) *api.ReplicationController {
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
func (k *Kubernetes) InitSvc(name string, service kobject.ServiceConfig) *api.Service {
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
func (k *Kubernetes) InitD(name string, service kobject.ServiceConfig, replicas int) *extensions.Deployment {
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
func (k *Kubernetes) InitDS(name string, service kobject.ServiceConfig) *extensions.DaemonSet {
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
func (k *Kubernetes) CreatePVC(name string, mode string) *api.PersistentVolumeClaim {
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
func (k *Kubernetes) ConfigPorts(name string, service kobject.ServiceConfig) []api.ContainerPort {
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
func (k *Kubernetes) ConfigServicePorts(name string, service kobject.ServiceConfig) []api.ServicePort {
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
func (k *Kubernetes) ConfigVolumes(name string, service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume, []*api.PersistentVolumeClaim) {
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
		pvc = append(pvc, k.CreatePVC(volumeName, mode))
	}
	return volumesMount, volumes, pvc
}

// Configure the environment variables.
func (k *Kubernetes) ConfigEnvs(name string, service kobject.ServiceConfig) []api.EnvVar {
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
func (k *Kubernetes) CreateKubernetesObjects(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions) []runtime.Object {
	var objects []runtime.Object

	if opt.CreateD {
		objects = append(objects, k.InitD(name, service, opt.Replicas))
	}
	if opt.CreateDS {
		objects = append(objects, k.InitDS(name, service))
	}
	if opt.CreateRC {
		objects = append(objects, k.InitRC(name, service, opt.Replicas))
	}

	return objects
}

// Transform maps komposeObject to k8s objects
// returns object that are already sorted in the way that Services are first
func (k *Kubernetes) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) []runtime.Object {
	// this will hold all the converted data
	var allobjects []runtime.Object

	for name, service := range komposeObject.ServiceConfigs {
		objects := k.CreateKubernetesObjects(name, service, opt)

		// If ports not provided in configuration we will not make service
		if k.PortsExist(name, service) {
			svc := k.CreateService(name, service, objects)
			objects = append(objects, svc)
		}

		k.UpdateKubernetesObjects(name, service, &objects)

		allobjects = append(allobjects, objects...)
	}
	// If docker-compose has a volumes_from directive it will be handled here
	k.VolumesFrom(&allobjects, komposeObject)
	// sort all object so Services are first
	k.SortServicesFirst(&allobjects)
	return allobjects
}

// Updates the given object with the given pod template update function and ObjectMeta update function
func (k *Kubernetes) UpdateController(obj runtime.Object, updateTemplate func(*api.PodTemplateSpec), updateMeta func(meta *api.ObjectMeta)) {
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

	fmt.Println("We are going to create Kubernetes Deployments, Services and PersistentVolumeClaims for your Dockerized application. \n" +
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
		case *api.PersistentVolumeClaim:
			_, err := client.PersistentVolumeClaims(namespace).Create(t)
			if err != nil {
				return err
			}
			logrus.Infof("Successfully created persistentVolumeClaim: %s", t.Name)
		}
	}
	fmt.Println("\nYour application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods,pvc' for details.")

	return nil
}

func (k *Kubernetes) Undeploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects := k.Transform(komposeObject, opt)

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
			//delete deployment
			rpDeployment, err := kubectl.ReaperFor(extensions.Kind("Deployment"), client)
			if err != nil {
				return err
			}
			//FIXME: gracePeriod is nil
			err = rpDeployment.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
			if err != nil {
				return err
			} else {
				logrus.Infof("Successfully deleted deployment: %s", t.Name)
			}
		case *api.Service:
			//delete svc
			rpService, err := kubectl.ReaperFor(api.Kind("Service"), client)
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
			err = client.PersistentVolumeClaims(namespace).Delete(t.Name)
			if err != nil {
				return err
			} else {
				logrus.Infof("Successfully deleted PersistentVolumeClaim: %s", t.Name)
			}
		}

	}
	return nil
}
