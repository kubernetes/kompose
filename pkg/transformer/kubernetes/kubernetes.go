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
	"os"

	"github.com/Sirupsen/logrus"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/skippbox/kompose/pkg/kobject"
	"github.com/skippbox/kompose/pkg/transformer"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/runtime"
)

type Kubernetes struct {
}

func (k *Kubernetes) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) (map[string][]byte, map[string][]byte, map[string][]byte, map[string][]byte, map[string][]byte, []string){
	mServices := make(map[string][]byte)
	mReplicationControllers := make(map[string][]byte)
	mDeployments := make(map[string][]byte)
	mDaemonSets := make(map[string][]byte)

	var svcnames []string

	for name, service := range komposeObject.ServiceConfigs {
		svcnames = append(svcnames, name)

		rc := transformer.InitRC(name, service, opt.Replicas)
		sc := transformer.InitSC(name, service)
		dc := transformer.InitDC(name, service, opt.Replicas)
		ds := transformer.InitDS(name, service)

		// Configure the environment variables.
		envs := transformer.ConfigEnvs(name, service)

		// Configure the container command.
		cmds := transformer.ConfigCommands(service)

		// Configure the container volumes.
		volumesMount, volumes := transformer.ConfigVolumes(service)

		// Configure the container ports.
		ports := transformer.ConfigPorts(name, service)

		// Configure the service ports.
		servicePorts := transformer.ConfigServicePorts(name, service)
		sc.Spec.Ports = servicePorts

		// Configure label
		labels := transformer.ConfigLabels(name)
		sc.ObjectMeta.Labels = labels

		// Configure annotations
		annotations := transformer.ConfigAnnotations(service)
		sc.ObjectMeta.Annotations = annotations

		// fillTemplate fills the pod template with the value calculated from config
		fillTemplate := func(template *api.PodTemplateSpec) {
			template.Spec.Containers[0].Env = envs
			template.Spec.Containers[0].Command = cmds
			template.Spec.Containers[0].WorkingDir = service.WorkingDir
			template.Spec.Containers[0].VolumeMounts = volumesMount
			template.Spec.Volumes = volumes
			// Configure the container privileged mode
			if service.Privileged == true {
				template.Spec.Containers[0].SecurityContext = &api.SecurityContext{
					Privileged: &service.Privileged,
				}
			}
			template.Spec.Containers[0].Ports = ports
			template.ObjectMeta.Labels = labels
			// Configure the container restart policy.
			switch service.Restart {
			case "", "always":
				template.Spec.RestartPolicy = api.RestartPolicyAlways
			case "no":
				template.Spec.RestartPolicy = api.RestartPolicyNever
			case "on-failure":
				template.Spec.RestartPolicy = api.RestartPolicyOnFailure
			default:
				logrus.Fatalf("Unknown restart policy %s for service %s", service.Restart, name)
			}
		}

		// fillObjectMeta fills the metadata with the value calculated from config
		fillObjectMeta := func(meta *api.ObjectMeta) {
			meta.Labels = labels
			meta.Annotations = annotations
		}

		// Update each supported controllers
		UpdateController(rc, fillTemplate, fillObjectMeta)
		UpdateController(dc, fillTemplate, fillObjectMeta)
		UpdateController(ds, fillTemplate, fillObjectMeta)

		// convert datarc to json / yaml
		datarc, err := transformer.TransformData(rc, opt.GenerateYaml)
		if err != nil {
			logrus.Fatalf(err.Error())
		}

		// convert datadc to json / yaml
		datadc, err := transformer.TransformData(dc, opt.GenerateYaml)
		if err != nil {
			logrus.Fatalf(err.Error())
		}

		// convert datads to json / yaml
		datads, err := transformer.TransformData(ds, opt.GenerateYaml)
		if err != nil {
			logrus.Fatalf(err.Error())
		}

		var datasvc []byte
		// If ports not provided in configuration we will not make service
		if len(ports) == 0 {
			logrus.Warningf("[%s] Service cannot be created because of missing port.", name)
		} else if len(ports) != 0 {
			// convert datasvc to json / yaml
			datasvc, err = transformer.TransformData(sc, opt.GenerateYaml)
			if err != nil {
				logrus.Fatalf(err.Error())
			}
		}

		mServices[name] = datasvc
		mReplicationControllers[name] = datarc
		mDeployments[name] = datadc
		mDaemonSets[name] = datads
	}

	return mServices, mDeployments, mDaemonSets, mReplicationControllers, nil, svcnames
}

// updateController updates the given object with the given pod template update function and ObjectMeta update function
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
	case *extensions.ReplicaSet:
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
