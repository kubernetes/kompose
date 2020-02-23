/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
	"github.com/spf13/pflag"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/fatih/structs"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	buildapi "github.com/openshift/origin/pkg/build/api"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	log "github.com/sirupsen/logrus"

	// install kubernetes api
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"

	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"

	//"k8s.io/kubernetes/pkg/controller/daemon"
	"sort"
	"strings"

	"path/filepath"

	"github.com/kubernetes/kompose/pkg/loader/compose"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/labels"

	utilflag "k8s.io/kubernetes/pkg/util/flag"
)

// Kubernetes implements Transformer interface and represents Kubernetes transformer
type Kubernetes struct {
	// the user provided options from the command line
	Opt kobject.ConvertOptions
}

// TIMEOUT is how long we'll wait for the termination of kubernetes resource to be successful
// used when undeploying resources from kubernetes
const TIMEOUT = 300

// PVCRequestSize (Persistent Volume Claim) has default size
const PVCRequestSize = "100Mi"

const (
	// DeploymentController is controller type for Deployment
	DeploymentController = "deployment"
	// DaemonSetController is controller type for DaemonSet
	DaemonSetController = "daemonset"
	// ReplicationController is controller type for  ReplicationController
	ReplicationController = "replicationcontroller"
)

// CheckUnsupportedKey checks if given komposeObject contains
// keys that are not supported by this transformer.
// list of all unsupported keys are stored in unsupportedKey variable
// returns list of TODO: ....
func (k *Kubernetes) CheckUnsupportedKey(komposeObject *kobject.KomposeObject, unsupportedKey map[string]bool) []string {
	// collect all keys found in project
	var keysFound []string

	for _, serviceConfig := range komposeObject.ServiceConfigs {
		// this reflection is used in check for empty arrays
		val := reflect.ValueOf(serviceConfig)
		s := structs.New(serviceConfig)

		for _, f := range s.Fields() {
			// Check if given key is among unsupported keys, and skip it if we already saw this key
			if alreadySaw, ok := unsupportedKey[f.Name()]; ok && !alreadySaw {

				if f.IsExported() && !f.IsZero() {
					// IsZero returns false for empty array/slice ([])
					// this check if field is Slice, and then it checks its size
					if field := val.FieldByName(f.Name()); field.Kind() == reflect.Slice {
						if field.Len() == 0 {
							// array is empty it doesn't matter if it is in unsupportedKey or not
							continue
						}
					}
					//get tag from kobject service configure
					tag := f.Tag(komposeObject.LoadedFrom)
					keysFound = append(keysFound, tag)
					unsupportedKey[f.Name()] = true
				}
			}
		}
	}
	return keysFound
}

// InitPodSpec creates the pod specification
func (k *Kubernetes) InitPodSpec(name string, image string, pullSecret string) api.PodSpec {

	if image == "" {
		image = name
	}

	pod := api.PodSpec{
		Containers: []api.Container{
			{
				Name:  name,
				Image: image,
			},
		},
	}
	if pullSecret != "" {
		pod.ImagePullSecrets = []api.LocalObjectReference{
			{
				Name: pullSecret,
			},
		}
	}
	return pod
}

//InitPodSpecWithConfigMap creates the pod specification
func (k *Kubernetes) InitPodSpecWithConfigMap(name string, image string, service kobject.ServiceConfig) api.PodSpec {
	var volumeMounts []api.VolumeMount
	var volumes []api.Volume

	log.Debugf("fuck config: %+v", service.Configs)

	for _, value := range service.Configs {
		cmVolName := FormatFileName(value.Source)
		target := value.Target
		if target == "" {
			// short syntax, = /<source>
			target = "/" + value.Source
		}
		subPath := filepath.Base(target)

		volSource := api.ConfigMapVolumeSource{}
		volSource.Name = cmVolName
		key, err := service.GetConfigMapKeyFromMeta(value.Source)
		if err != nil {
			log.Warnf("cannot parse config %s , %s", value.Source, err.Error())
			// mostly it's external
			continue
		}
		volSource.Items = []api.KeyToPath{{
			Key:  key,
			Path: subPath,
		}}

		if value.Mode != nil {
			tmpMode := int32(*value.Mode)
			volSource.DefaultMode = &tmpMode
		}

		cmVol := api.Volume{
			Name:         cmVolName,
			VolumeSource: api.VolumeSource{ConfigMap: &volSource},
		}

		volumeMounts = append(volumeMounts,
			api.VolumeMount{
				Name:      cmVolName,
				MountPath: target,
				SubPath:   subPath,
			})
		volumes = append(volumes, cmVol)

	}

	pod := api.PodSpec{
		Containers: []api.Container{
			{
				Name:         name,
				Image:        image,
				VolumeMounts: volumeMounts,
			},
		},
		Volumes: volumes,
	}
	return pod
}

// InitRC initializes Kubernetes ReplicationController object
func (k *Kubernetes) InitRC(name string, service kobject.ServiceConfig, replicas int) *api.ReplicationController {
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: int32(replicas),
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: transformer.ConfigLabels(name),
				},
				Spec: k.InitPodSpec(name, service.Image, service.ImagePullSecret),
			},
		},
	}
	return rc
}

// InitSvc initializes Kubernetes Service object
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

// InitConfigMapForEnv initializes a ConfigMap object
func (k *Kubernetes) InitConfigMapForEnv(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions, envFile string) *api.ConfigMap {

	envs, err := GetEnvsFromFile(envFile, opt)
	if err != nil {
		log.Fatalf("Unable to retrieve env file: %s", err)
	}

	// Remove root pathing
	// replace all other slashes / periods
	envName := FormatEnvName(envFile)

	// In order to differentiate files, we append to the name and remove '.env' if applicable from the file name
	configMap := &api.ConfigMap{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   envName,
			Labels: transformer.ConfigLabels(name + "-" + envName),
		},
		Data: envs,
	}

	return configMap
}

// IntiConfigMapFromFileOrDir will create a configmap from dir or file
// usage:
//   1. volume
func (k *Kubernetes) IntiConfigMapFromFileOrDir(name, cmName, filePath string, service kobject.ServiceConfig) (*api.ConfigMap, error) {
	configMap := &api.ConfigMap{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   cmName,
			Labels: transformer.ConfigLabels(name),
		},
	}
	dataMap := make(map[string]string)

	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, err := ioutil.ReadDir(filePath)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !file.IsDir() {
				log.Debugf("Read file to ConfigMap: %s", file.Name())
				data, err := GetContentFromFile(filePath + "/" + file.Name())
				if err != nil {
					return nil, err
				}
				dataMap[file.Name()] = data
			}
		}
		configMap.Data = dataMap

	case mode.IsRegular():
		// do file stuff
		configMap = k.InitConfigMapFromFile(name, service, filePath)
		configMap.Name = cmName
		configMap.Annotations = map[string]string{
			"use-subpath": "true",
		}
	}

	return configMap, nil
}

// useSubPathMount check if a configmap should be mounted as subpath
// in this situation, this configmap will only contains 1 key in data
func useSubPathMount(cm *api.ConfigMap) bool {
	if cm.Annotations == nil {
		return false
	}
	if cm.Annotations["use-subpath"] != "true" {
		return false
	}
	return true
}

//InitConfigMapFromFile initializes a ConfigMap object
func (k *Kubernetes) InitConfigMapFromFile(name string, service kobject.ServiceConfig, fileName string) *api.ConfigMap {
	content, err := GetContentFromFile(fileName)
	if err != nil {
		log.Fatalf("Unable to retrieve file: %s", err)
	}

	dataMap := make(map[string]string)
	dataMap[filepath.Base(fileName)] = content

	configMapName := ""
	for key, tmpConfig := range service.ConfigsMetaData {
		if tmpConfig.File == fileName {
			configMapName = key
		}
	}
	configMap := &api.ConfigMap{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   FormatFileName(configMapName),
			Labels: transformer.ConfigLabels(name),
		},
		Data: dataMap,
	}
	return configMap
}

// InitD initializes Kubernetes Deployment object
func (k *Kubernetes) InitD(name string, service kobject.ServiceConfig, replicas int) *extensions.Deployment {

	var podSpec api.PodSpec
	if len(service.Configs) > 0 {
		podSpec = k.InitPodSpecWithConfigMap(name, service.Image, service)
	} else {
		podSpec = k.InitPodSpec(name, service.Image, service.ImagePullSecret)
	}

	dc := &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigAllLabels(name, &service),
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32(replicas),
			Selector: &unversioned.LabelSelector{
				MatchLabels: transformer.ConfigLabels(name),
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					//Labels: transformer.ConfigLabels(name),
					Annotations: transformer.ConfigAnnotations(service),
				},
				Spec: podSpec,
			},
		},
	}
	dc.Spec.Template.Labels = transformer.ConfigLabels(name)

	update := service.GetKubernetesUpdateStrategy()
	if update != nil {
		dc.Spec.Strategy = extensions.DeploymentStrategy{
			Type:          extensions.RollingUpdateDeploymentStrategyType,
			RollingUpdate: update,
		}
		log.Debugf("Set deployment '%s' rolling update: MaxSurge: %s, MaxUnavailable: %s", name, update.MaxSurge.String(), update.MaxUnavailable.String())
	}

	return dc
}

// InitDS initializes Kubernetes DaemonSet object
func (k *Kubernetes) InitDS(name string, service kobject.ServiceConfig) *extensions.DaemonSet {
	ds := &extensions.DaemonSet{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigAllLabels(name, &service),
		},
		Spec: extensions.DaemonSetSpec{
			Template: api.PodTemplateSpec{
				Spec: k.InitPodSpec(name, service.Image, service.ImagePullSecret),
			},
		},
	}
	return ds
}

func (k *Kubernetes) initIngress(name string, service kobject.ServiceConfig, port int32) *extensions.Ingress {

	hosts := regexp.MustCompile("[ ,]*,[ ,]*").Split(service.ExposeService, -1)

	ingress := &extensions.Ingress{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: extensions.IngressSpec{
			Rules: make([]extensions.IngressRule, len(hosts)),
		},
	}

	for i, host := range hosts {
		host, p := transformer.ParseIngressPath(host)
		ingress.Spec.Rules[i] = extensions.IngressRule{
			IngressRuleValue: extensions.IngressRuleValue{
				HTTP: &extensions.HTTPIngressRuleValue{
					Paths: []extensions.HTTPIngressPath{
						{
							Path: p,
							Backend: extensions.IngressBackend{
								ServiceName: name,
								ServicePort: intstr.IntOrString{
									IntVal: port,
								},
							},
						},
					},
				},
			},
		}
		if host != "true" {
			ingress.Spec.Rules[i].Host = host
		}
	}

	if service.ExposeServiceTLS != "" {
		ingress.Spec.TLS = []extensions.IngressTLS{
			{
				Hosts:      hosts,
				SecretName: service.ExposeServiceTLS,
			},
		}
	}

	return ingress
}

// CreateSecrets create secrets
func (k *Kubernetes) CreateSecrets(komposeObject kobject.KomposeObject) ([]*api.Secret, error) {
	var objects []*api.Secret
	for name, config := range komposeObject.Secrets {
		if config.File != "" {
			dataString, err := GetContentFromFile(config.File)
			if err != nil {
				log.Fatal("unable to read secret from file: ", config.File)
				return nil, err
			}
			data := []byte(dataString)
			secret := &api.Secret{
				TypeMeta: unversioned.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: api.ObjectMeta{
					Name:   name,
					Labels: transformer.ConfigLabels(name),
				},
				Type: api.SecretTypeOpaque,
				Data: map[string][]byte{name: data},
			}
			objects = append(objects, secret)
		} else {
			log.Warnf("External secrets %s is not currently supported - ignoring", name)
		}
	}
	return objects, nil

}

// CreatePVC initializes PersistentVolumeClaim
func (k *Kubernetes) CreatePVC(name string, mode string, size string, selectorValue string) (*api.PersistentVolumeClaim, error) {
	volSize, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrap(err, "resource.ParseQuantity failed, Error parsing size")
	}

	pvc := &api.PersistentVolumeClaim{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: api.PersistentVolumeClaimSpec{
			Resources: api.ResourceRequirements{
				Requests: api.ResourceList{
					api.ResourceStorage: volSize,
				},
			},
		},
	}

	if len(selectorValue) > 0 {
		pvc.Spec.Selector = &unversioned.LabelSelector{
			MatchLabels: transformer.ConfigLabels(selectorValue),
		}
	}

	if mode == "ro" {
		pvc.Spec.AccessModes = []api.PersistentVolumeAccessMode{api.ReadOnlyMany}
	} else {
		pvc.Spec.AccessModes = []api.PersistentVolumeAccessMode{api.ReadWriteOnce}
	}
	return pvc, nil
}

// ConfigPorts configures the container ports.
func (k *Kubernetes) ConfigPorts(name string, service kobject.ServiceConfig) []api.ContainerPort {
	ports := []api.ContainerPort{}
	exist := map[string]bool{}
	for _, port := range service.Port {
		// temp use as an id
		if exist[string(port.ContainerPort)+string(port.Protocol)] {
			continue
		}
		// If the default is already TCP, no need to include it.
		if port.Protocol == api.ProtocolTCP {
			ports = append(ports, api.ContainerPort{
				ContainerPort: port.ContainerPort,
				HostIP:        port.HostIP,
			})
		} else {
			ports = append(ports, api.ContainerPort{
				ContainerPort: port.ContainerPort,
				Protocol:      port.Protocol,
				HostIP:        port.HostIP,
			})
		}
		exist[string(port.ContainerPort)+string(port.Protocol)] = true

	}

	return ports
}

// ConfigServicePorts configure the container service ports.
func (k *Kubernetes) ConfigServicePorts(name string, service kobject.ServiceConfig) []api.ServicePort {
	servicePorts := []api.ServicePort{}
	seenPorts := make(map[int]struct{}, len(service.Port))

	var servicePort api.ServicePort
	for _, port := range service.Port {
		if port.HostPort == 0 {
			port.HostPort = port.ContainerPort
		}

		var targetPort intstr.IntOrString
		targetPort.IntVal = port.ContainerPort
		targetPort.StrVal = strconv.Itoa(int(port.ContainerPort))

		// decide the name based on whether we saw this port before
		name := strconv.Itoa(int(port.HostPort))
		if _, ok := seenPorts[int(port.HostPort)]; ok {
			// https://github.com/kubernetes/kubernetes/issues/2995
			if service.ServiceType == string(api.ServiceTypeLoadBalancer) {
				log.Fatalf("Service %s of type LoadBalancer cannot use TCP and UDP for the same port", name)
			}
			name = fmt.Sprintf("%s-%s", name, strings.ToLower(string(port.Protocol)))
		}

		servicePort = api.ServicePort{
			Name:       name,
			Port:       port.HostPort,
			TargetPort: targetPort,
		}

		if service.ServiceType == string(api.ServiceTypeNodePort) && service.NodePortPort != 0 {
			servicePort.NodePort = service.NodePortPort
		}

		// If the default is already TCP, no need to include it.
		if port.Protocol != api.ProtocolTCP {
			servicePort.Protocol = port.Protocol
		}

		servicePorts = append(servicePorts, servicePort)
		seenPorts[int(port.HostPort)] = struct{}{}
	}
	return servicePorts
}

//ConfigCapabilities configure POSIX capabilities that can be added or removed to a container
func (k *Kubernetes) ConfigCapabilities(service kobject.ServiceConfig) *api.Capabilities {
	capsAdd := []api.Capability{}
	capsDrop := []api.Capability{}
	for _, capAdd := range service.CapAdd {
		capsAdd = append(capsAdd, api.Capability(capAdd))
	}
	for _, capDrop := range service.CapDrop {
		capsDrop = append(capsDrop, api.Capability(capDrop))
	}
	return &api.Capabilities{
		Add:  capsAdd,
		Drop: capsDrop,
	}
}

// ConfigTmpfs configure the tmpfs.
func (k *Kubernetes) ConfigTmpfs(name string, service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume) {
	//initializing volumemounts and volumes
	volumeMounts := []api.VolumeMount{}
	volumes := []api.Volume{}

	for index, volume := range service.TmpFs {
		//naming volumes if multiple tmpfs are provided
		volumeName := fmt.Sprintf("%s-tmpfs%d", name, index)
		volume = strings.Split(volume, ":")[0]
		// create a new volume mount object and append to list
		volMount := api.VolumeMount{
			Name:      volumeName,
			MountPath: volume,
		}
		volumeMounts = append(volumeMounts, volMount)

		//create tmpfs specific empty volumes
		volSource := k.ConfigEmptyVolumeSource("tmpfs")

		// create a new volume object using the volsource and add to list
		vol := api.Volume{
			Name:         volumeName,
			VolumeSource: *volSource,
		}
		volumes = append(volumes, vol)
	}
	return volumeMounts, volumes
}

// ConfigSecretVolumes config volumes from secret.
// Link: https://docs.docker.com/compose/compose-file/#secrets
// In kubernetes' Secret resource, it has a data structure like a map[string]bytes, every key will act like the file name
// when mount to a container. This is the part that missing in compose. So we will create a single key secret from compose
// config and the key's name will be the secret's name, it's value is the file content.
// compose'secret can only be mounted at `/run/secrets`, so we will hardcoded this.
func (k *Kubernetes) ConfigSecretVolumes(name string, service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume) {
	var volumeMounts []api.VolumeMount
	var volumes []api.Volume
	if len(service.Secrets) > 0 {
		for _, secretConfig := range service.Secrets {
			if secretConfig.UID != "" {
				log.Warnf("Ignore pid in secrets for service: %s", name)
			}
			if secretConfig.GID != "" {
				log.Warnf("Ignore gid in secrets for service: %s", name)
			}

			var itemPath string // should be the filename
			var mountPath = ""  // should be the directory
			// if is used the short-syntax
			if secretConfig.Target == "" {
				// the secret path (mountPath) should be inside the default directory /run/secrets
				mountPath = "/run/secrets/" + secretConfig.Source
				// the itemPath should be the source itself
				itemPath = secretConfig.Source
			} else {
				// if is the long-syntax, i should get the last part of path and consider it the filename
				pathSplitted := strings.Split(secretConfig.Target, "/")
				lastPart := pathSplitted[len(pathSplitted)-1]

				// if the filename (lastPart) and the target is the same
				if lastPart == secretConfig.Target {
					// the secret path should be the source (it need to be inside a directory and only the filename was given)
					mountPath = secretConfig.Source
				} else {
					// should then get the target without the filename (lastPart)
					mountPath = mountPath + strings.TrimSuffix(secretConfig.Target, "/"+lastPart) // menos ultima parte
				}

				// if the target isn't absolute path
				if strings.HasPrefix(secretConfig.Target, "/") == false {
					// concat the default secret directory
					mountPath = "/run/secrets/" + mountPath
				}

				itemPath = lastPart
			}

			volSource := api.VolumeSource{
				Secret: &api.SecretVolumeSource{
					SecretName: secretConfig.Source,
					Items: []api.KeyToPath{{
						Key:  secretConfig.Source,
						Path: itemPath,
					}},
				},
			}

			if secretConfig.Mode != nil {
				mode := cast.ToInt32(*secretConfig.Mode)
				volSource.Secret.DefaultMode = &mode
			}

			vol := api.Volume{
				Name:         secretConfig.Source,
				VolumeSource: volSource,
			}
			volumes = append(volumes, vol)

			volMount := api.VolumeMount{
				Name:      vol.Name,
				MountPath: mountPath,
			}
			volumeMounts = append(volumeMounts, volMount)
		}
	}
	return volumeMounts, volumes
}

// ConfigVolumes configure the container volumes.
func (k *Kubernetes) ConfigVolumes(name string, service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume, []*api.PersistentVolumeClaim, []*api.ConfigMap, error) {
	volumeMounts := []api.VolumeMount{}
	volumes := []api.Volume{}
	var PVCs []*api.PersistentVolumeClaim
	var cms []*api.ConfigMap
	var volumeName string

	// Set a var based on if the user wants to use empty volumes
	// as opposed to persistent volumes and volume claims
	useEmptyVolumes := k.Opt.EmptyVols
	useHostPath := k.Opt.Volumes == "hostPath"
	useConfigMap := k.Opt.Volumes == "configMap"

	if k.Opt.Volumes == "emptyDir" {
		useEmptyVolumes = true
	}

	// config volumes from secret if present
	secretsVolumeMounts, secretsVolumes := k.ConfigSecretVolumes(name, service)
	volumeMounts = append(volumeMounts, secretsVolumeMounts...)
	volumes = append(volumes, secretsVolumes...)

	var count int
	//iterating over array of `Vols` struct as it contains all necessary information about volumes
	for _, volume := range service.Volumes {

		// check if ro/rw mode is defined, default rw
		readonly := len(volume.Mode) > 0 && volume.Mode == "ro"

		if volume.VolumeName == "" {
			if useEmptyVolumes {
				volumeName = strings.Replace(volume.PVCName, "claim", "empty", 1)
			} else if useHostPath {
				volumeName = strings.Replace(volume.PVCName, "claim", "hostpath", 1)
			} else if useConfigMap {
				volumeName = strings.Replace(volume.PVCName, "claim", "cm", 1)
			} else {
				volumeName = volume.PVCName
			}
			count++
		} else {
			volumeName = volume.VolumeName
		}
		volMount := api.VolumeMount{
			Name:      volumeName,
			ReadOnly:  readonly,
			MountPath: volume.Container,
		}

		// Get a volume source based on the type of volume we are using
		// For PVC we will also create a PVC object and add to list
		var volsource *api.VolumeSource

		if useEmptyVolumes {
			volsource = k.ConfigEmptyVolumeSource("volume")
		} else if useHostPath {
			source, err := k.ConfigHostPathVolumeSource(volume.Host)
			if err != nil {
				return nil, nil, nil, nil, errors.Wrap(err, "k.ConfigHostPathVolumeSource failed")
			}
			volsource = source
		} else if useConfigMap {
			log.Debugf("Use configmap volume")

			if cm, err := k.IntiConfigMapFromFileOrDir(name, volumeName, volume.Host, service); err != nil {
				return nil, nil, nil, nil, err
			} else {
				cms = append(cms, cm)
				volsource = k.ConfigConfigMapVolumeSource(volumeName, volume.Container, cm)

				if useSubPathMount(cm) {
					volMount.SubPath = volsource.ConfigMap.Items[0].Path
				}
			}

		} else {
			volsource = k.ConfigPVCVolumeSource(volumeName, readonly)
			if volume.VFrom == "" {
				defaultSize := PVCRequestSize

				if len(volume.PVCSize) > 0 {
					defaultSize = volume.PVCSize
				} else {
					for key, value := range service.Labels {
						if key == "kompose.volume.size" {
							defaultSize = value
						}
					}
				}

				createdPVC, err := k.CreatePVC(volumeName, volume.Mode, defaultSize, volume.SelectorValue)

				if err != nil {
					return nil, nil, nil, nil, errors.Wrap(err, "k.CreatePVC failed")
				}

				PVCs = append(PVCs, createdPVC)
			}

		}
		volumeMounts = append(volumeMounts, volMount)

		// create a new volume object using the volsource and add to list
		vol := api.Volume{
			Name:         volumeName,
			VolumeSource: *volsource,
		}
		volumes = append(volumes, vol)

		if len(volume.Host) > 0 && (!useHostPath && !useConfigMap) {
			log.Warningf("Volume mount on the host %q isn't supported - ignoring path on the host", volume.Host)
		}

	}

	return volumeMounts, volumes, PVCs, cms, nil
}

// ConfigEmptyVolumeSource is helper function to create an EmptyDir api.VolumeSource
//either for Tmpfs or for emptyvolumes
func (k *Kubernetes) ConfigEmptyVolumeSource(key string) *api.VolumeSource {
	//if key is tmpfs
	if key == "tmpfs" {
		return &api.VolumeSource{
			EmptyDir: &api.EmptyDirVolumeSource{Medium: api.StorageMediumMemory},
		}

	}

	//if key is volume
	return &api.VolumeSource{
		EmptyDir: &api.EmptyDirVolumeSource{},
	}

}

// ConfigHostPathVolumeSource config a configmap to use as volume source
func (k *Kubernetes) ConfigConfigMapVolumeSource(cmName string, targetPath string, cm *api.ConfigMap) *api.VolumeSource {
	s := api.ConfigMapVolumeSource{}
	s.Name = cmName
	if useSubPathMount(cm) {
		var keys []string
		for k := range cm.Data {
			keys = append(keys, k)
		}
		key := keys[0]
		_, p := path.Split(targetPath)
		s.Items = []api.KeyToPath{
			{
				Key:  key,
				Path: p,
			},
		}
	}
	return &api.VolumeSource{
		ConfigMap: &s,
	}

}

// ConfigHostPathVolumeSource is a helper function to create a HostPath api.VolumeSource
func (k *Kubernetes) ConfigHostPathVolumeSource(path string) (*api.VolumeSource, error) {
	dir, err := transformer.GetComposeFileDir(k.Opt.InputFiles)
	if err != nil {
		return nil, err
	}
	absPath := path
	if !filepath.IsAbs(path) {
		absPath = filepath.Join(dir, path)
	}

	return &api.VolumeSource{
		HostPath: &api.HostPathVolumeSource{Path: absPath},
	}, nil
}

// ConfigPVCVolumeSource is helper function to create an api.VolumeSource with a PVC
func (k *Kubernetes) ConfigPVCVolumeSource(name string, readonly bool) *api.VolumeSource {
	return &api.VolumeSource{
		PersistentVolumeClaim: &api.PersistentVolumeClaimVolumeSource{
			ClaimName: name,
			ReadOnly:  readonly,
		},
	}
}

// ConfigEnvs configures the environment variables.
func (k *Kubernetes) ConfigEnvs(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions) ([]api.EnvVar, error) {

	envs := transformer.EnvSort{}

	keysFromEnvFile := make(map[string]bool)

	// If there is an env_file, use ConfigMaps and ignore the environment variables
	// already specified

	if len(service.EnvFile) > 0 {

		// Load each env_file

		for _, file := range service.EnvFile {

			envName := FormatEnvName(file)

			// Load environment variables from file
			envLoad, err := GetEnvsFromFile(file, opt)
			if err != nil {
				return envs, errors.Wrap(err, "Unable to read env_file")
			}

			// Add configMapKeyRef to each environment variable
			for k := range envLoad {
				envs = append(envs, api.EnvVar{
					Name: k,
					ValueFrom: &api.EnvVarSource{
						ConfigMapKeyRef: &api.ConfigMapKeySelector{
							LocalObjectReference: api.LocalObjectReference{
								Name: envName,
							},
							Key: k,
						}},
				})
				keysFromEnvFile[k] = true
			}
		}
	}

	// Load up the environment variables
	for _, v := range service.Environment {
		if !keysFromEnvFile[v.Name] {
			envs = append(envs, api.EnvVar{
				Name:  v.Name,
				Value: v.Value,
			})
		}

	}

	// Stable sorts data while keeping the original order of equal elements
	// we need this because envs are not populated in any random order
	// this sorting ensures they are populated in a particular order
	sort.Stable(envs)
	return envs, nil
}

// CreateKubernetesObjects generates a Kubernetes artifact for each input type service
func (k *Kubernetes) CreateKubernetesObjects(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions) []runtime.Object {
	var objects []runtime.Object
	var replica int

	if opt.IsReplicaSetFlag || service.Replicas == 0 {
		replica = opt.Replicas
	} else {
		replica = service.Replicas
	}

	// Check to see if Docker Compose v3 Deploy.Mode has been set to "global"
	if service.DeployMode == "global" {
		//default use daemonset
		if opt.Controller == "" {
			opt.CreateD = false
			opt.CreateDS = true
		} else if opt.Controller != "daemonset" {
			log.Warnf("Global deploy mode service is best converted to daemonset, now it convert to %s", opt.Controller)
		}

	}

	//Resolve labels first
	if val, ok := service.Labels[compose.LabelControllerType]; ok {
		opt.CreateD = false
		opt.CreateDS = false
		opt.CreateRC = false
		if opt.Controller != "" {
			log.Warnf("Use label %s type %s for service %s, ignore %s flags", compose.LabelControllerType, val, name, opt.Controller)
		}
		opt.Controller = val
	}

	if len(service.Configs) > 0 {
		objects = k.createConfigMapFromComposeConfig(name, opt, service, objects)
	}

	if opt.CreateD || opt.Controller == DeploymentController {
		objects = append(objects, k.InitD(name, service, replica))
	}

	if opt.CreateDS || opt.Controller == DaemonSetController {
		objects = append(objects, k.InitDS(name, service))
	}
	if opt.CreateRC || opt.Controller == ReplicationController {
		objects = append(objects, k.InitRC(name, service, replica))
	}

	if len(service.EnvFile) > 0 {
		for _, envFile := range service.EnvFile {
			configMap := k.InitConfigMapForEnv(name, service, opt, envFile)
			objects = append(objects, configMap)
		}
	}

	return objects
}

func (k *Kubernetes) createConfigMapFromComposeConfig(name string, opt kobject.ConvertOptions, service kobject.ServiceConfig, objects []runtime.Object) []runtime.Object {
	for _, config := range service.Configs {
		currentConfigName := config.Source
		currentConfigObj := service.ConfigsMetaData[currentConfigName]
		if currentConfigObj.External.External {
			continue
		}
		currentFileName := currentConfigObj.File
		configMap := k.InitConfigMapFromFile(name, service, currentFileName)
		objects = append(objects, configMap)
	}
	return objects
}

// InitPod initializes Kubernetes Pod object
func (k *Kubernetes) InitPod(name string, service kobject.ServiceConfig) *api.Pod {
	pod := api.Pod{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: k.InitPodSpec(name, service.Image, service.ImagePullSecret),
	}
	return &pod
}

// CreateNetworkPolicy initializes Network policy
func (k *Kubernetes) CreateNetworkPolicy(name string, networkName string) (*extensions.NetworkPolicy, error) {

	str := "true"
	np := &extensions.NetworkPolicy{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "NetworkPolicy",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: networkName,
			//Labels: transformer.ConfigLabels(name)(name),
		},
		Spec: extensions.NetworkPolicySpec{
			PodSelector: unversioned.LabelSelector{
				MatchLabels: map[string]string{"io.kompose.network/" + networkName: str},
			},
			Ingress: []extensions.NetworkPolicyIngressRule{{
				From: []extensions.NetworkPolicyPeer{{
					PodSelector: &unversioned.LabelSelector{
						MatchLabels: map[string]string{"io.kompose.network/" + networkName: str},
					},
				}},
			}},
		},
	}

	return np, nil
}

// Transform maps komposeObject to k8s objects
// returns object that are already sorted in the way that Services are first
func (k *Kubernetes) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) ([]runtime.Object, error) {

	// this will hold all the converted data
	var allobjects []runtime.Object

	if komposeObject.Secrets != nil {
		secrets, err := k.CreateSecrets(komposeObject)
		if err != nil {
			return nil, errors.Wrapf(err, "Unable to create Secret resource")
		}
		for _, item := range secrets {
			allobjects = append(allobjects, item)
		}
	}

	sortedKeys := SortedKeys(komposeObject)
	for _, name := range sortedKeys {
		service := komposeObject.ServiceConfigs[name]
		var objects []runtime.Object

		// Must build the images before conversion (got to add service.Image in case 'image' key isn't provided
		// Check that --build is set to true
		// Check to see if there is an InputFile (required!) before we build the container
		// Check that there's actually a Build key
		// Lastly, we must have an Image name to continue
		if opt.Build == "local" && opt.InputFiles != nil && service.Build != "" {

			// If there's no "image" key, use the name of the container that's built
			if service.Image == "" {
				service.Image = name
			}

			if service.Image == "" {
				return nil, fmt.Errorf("image key required within build parameters in order to build and push service '%s'", name)
			}

			log.Infof("Build key detected. Attempting to build image '%s'", service.Image)

			// Build the image!
			err := transformer.BuildDockerImage(service, name)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to build Docker image for service %v", name)
			}

			// Push the built image to the repo!
			if opt.PushImage {
				log.Infof("Push image enabled. Attempting to push image '%s'", service.Image)
				err = transformer.PushDockerImage(service, name)
				if err != nil {
					return nil, errors.Wrapf(err, "Unable to push Docker image for service %v", name)
				}
			}
		}

		// Generate pod only and nothing more
		if service.Restart == "no" || service.Restart == "on-failure" {
			// Error out if Controller Object is specified with restart: 'on-failure'
			if opt.IsDeploymentFlag || opt.IsDaemonSetFlag || opt.IsReplicationControllerFlag {
				return nil, errors.New("Controller object cannot be specified with restart: 'on-failure'")
			}
			pod := k.InitPod(name, service)
			objects = append(objects, pod)
		} else {
			objects = k.CreateKubernetesObjects(name, service, opt)
		}

		if k.PortsExist(service) {
			svc := k.CreateService(name, service, objects)
			objects = append(objects, svc)

			if service.ExposeService != "" {
				objects = append(objects, k.initIngress(name, service, svc.Spec.Ports[0].Port))
			}
		} else {
			if service.ServiceType == "Headless" {
				svc := k.CreateHeadlessService(name, service, objects)
				objects = append(objects, svc)
			}
		}

		err := k.UpdateKubernetesObjects(name, service, opt, &objects)
		if err != nil {
			return nil, errors.Wrap(err, "Error transforming Kubernetes objects")
		}

		if len(service.Network) > 0 {

			for _, net := range service.Network {

				log.Infof("Network %s is detected at Source, shall be converted to equivalent NetworkPolicy at Destination", net)
				np, err := k.CreateNetworkPolicy(name, net)

				if err != nil {
					return nil, errors.Wrapf(err, "Unable to create Network Policy for network %v for service %v", net, name)
				}
				objects = append(objects, np)

			}

		}

		allobjects = append(allobjects, objects...)

	}

	// sort all object so Services are first
	k.SortServicesFirst(&allobjects)
	k.RemoveDupObjects(&allobjects)
	k.FixWorkloadVersion(&allobjects)

	return allobjects, nil
}

// UpdateController updates the given object with the given pod template update function and ObjectMeta update function
func (k *Kubernetes) UpdateController(obj runtime.Object, updateTemplate func(*api.PodTemplateSpec) error, updateMeta func(meta *api.ObjectMeta)) (err error) {
	switch t := obj.(type) {
	case *api.ReplicationController:
		if t.Spec.Template == nil {
			t.Spec.Template = &api.PodTemplateSpec{}
		}
		err = updateTemplate(t.Spec.Template)
		if err != nil {
			return errors.Wrap(err, "updateTemplate failed")
		}
		updateMeta(&t.ObjectMeta)
	case *extensions.Deployment:
		err = updateTemplate(&t.Spec.Template)
		if err != nil {
			return errors.Wrap(err, "updateTemplate failed")
		}
		updateMeta(&t.ObjectMeta)
	case *extensions.DaemonSet:
		err = updateTemplate(&t.Spec.Template)
		if err != nil {
			return errors.Wrap(err, "updateTemplate failed")
		}
		updateMeta(&t.ObjectMeta)
	case *deployapi.DeploymentConfig:
		err = updateTemplate(t.Spec.Template)
		if err != nil {
			return errors.Wrap(err, "updateTemplate failed")
		}
		updateMeta(&t.ObjectMeta)
	case *api.Pod:
		p := api.PodTemplateSpec{
			ObjectMeta: t.ObjectMeta,
			Spec:       t.Spec,
		}
		err = updateTemplate(&p)
		if err != nil {
			return errors.Wrap(err, "updateTemplate failed")
		}
		t.Spec = p.Spec
		t.ObjectMeta = p.ObjectMeta
	case *buildapi.BuildConfig:
		updateMeta(&t.ObjectMeta)
	}
	return nil
}

// DefaultClientConfig get default client config.
// This function is copied from library , we just overrides the apiserver url
func (k *Kubernetes) DefaultClientConfig(flags *pflag.FlagSet) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	flags.StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")

	clusterDefaults := clientcmd.ClusterDefaults
	clusterDefaults.Server = "https://127.0.0.1:6443"
	if k.Opt.Server != "" {
		clusterDefaults.Server = k.Opt.Server
	}

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clusterDefaults}

	flagNames := clientcmd.RecommendedConfigOverrideFlags("")
	// short flagnames are disabled by default.  These are here for compatibility with existing scripts
	flagNames.ClusterOverrideFlags.APIServer.ShortName = "s"

	clientcmd.BindOverrideFlags(overrides, flags, flagNames)
	clientConfig := clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, os.Stdin)

	return clientConfig
}

// GetKubernetesClient creates the k8s Client, returns k8s client and namespace
func (k *Kubernetes) GetKubernetesClient() (*client.Client, string, error) {

	// generate a new client config
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	flags.SetNormalizeFunc(utilflag.WarnWordSepNormalizeFunc) // Warn for "_" flags
	oc := k.DefaultClientConfig(flags)

	// initialize Kubernetes client
	factory := cmdutil.NewFactory(oc)
	clientConfig, err := factory.ClientConfig()
	if err != nil {
		return nil, "", err
	}
	client := client.NewOrDie(clientConfig)

	// get namespace from config
	namespace, _, err := factory.DefaultNamespace()
	if err != nil {
		return nil, "", err
	}
	return client, namespace, nil
}

// Deploy submits deployment and svc to k8s endpoint
func (k *Kubernetes) Deploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects, err := k.Transform(komposeObject, opt)

	if err != nil {
		return errors.Wrap(err, "k.Transform failed")
	}

	if opt.StoreManifest {
		log.Info("Store manifest to disk")
		if err := PrintList(objects, opt); err != nil {
			return errors.Wrap(err, "Store manifest failed")
		}
	}

	pvcStr := " "
	if !opt.EmptyVols || opt.Volumes != "emptyDir" {
		pvcStr = " and PersistentVolumeClaims "
	}
	log.Info("We are going to create Kubernetes Deployments, Services" + pvcStr + "for your Dockerized application. " +
		"If you need different kind of resources, use the 'kompose convert' and 'kubectl create -f' commands instead. \n")

	client, ns, err := k.GetKubernetesClient()
	namespace := ns
	if opt.IsNamespaceFlag {
		namespace = opt.Namespace
	}
	if err != nil {
		return err
	}

	pvcCreatedSet := make(map[string]bool)

	log.Infof("Deploying application in %q namespace", namespace)

	for _, v := range objects {
		switch t := v.(type) {
		case *extensions.Deployment:
			_, err := client.Deployments(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Deployment: %s", t.Name)

		case *extensions.DaemonSet:
			_, err := client.DaemonSets(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created DaemonSet: %s", t.Name)

		case *api.ReplicationController:
			_, err := client.ReplicationControllers(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created ReplicationController: %s", t.Name)

		case *api.Service:
			_, err := client.Services(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Service: %s", t.Name)
		case *api.PersistentVolumeClaim:
			if pvcCreatedSet[t.Name] {
				log.Infof("Skip creation of PersistentVolumeClaim as it is already created: %s", t.Name)
			} else {
				_, err := client.PersistentVolumeClaims(namespace).Create(t)
				if err != nil {
					return err
				}
				pvcCreatedSet[t.Name] = true
				storage := t.Spec.Resources.Requests[api.ResourceStorage]
				capacity := storage.String()
				log.Infof("Successfully created PersistentVolumeClaim: %s of size %s. If your cluster has dynamic storage provisioning, you don't have to do anything. Otherwise you have to create PersistentVolume to make PVC work", t.Name, capacity)
			}
		case *extensions.Ingress:
			_, err := client.Ingress(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Ingress: %s", t.Name)
		case *api.Pod:
			_, err := client.Pods(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Pod: %s", t.Name)
		case *api.ConfigMap:
			_, err := client.ConfigMaps(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Config Map: %s", t.Name)
		}
	}

	if !opt.EmptyVols || opt.Volumes != "emptyDir" {
		pvcStr = ",pvc"
	} else {
		pvcStr = ""
	}
	fmt.Println("\nYour application has been deployed to Kubernetes. You can run 'kubectl get deployment,svc,pods" + pvcStr + "' for details.")

	return nil
}

// Undeploy deletes deployed objects from Kubernetes cluster
func (k *Kubernetes) Undeploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) []error {
	var errorList []error
	//Convert komposeObject
	objects, err := k.Transform(komposeObject, opt)
	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}

	client, ns, err := k.GetKubernetesClient()
	namespace := ns
	if opt.IsNamespaceFlag {
		namespace = opt.Namespace
	}

	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}

	log.Infof("Deleting application in %q namespace", namespace)

	for _, v := range objects {
		label := labels.SelectorFromSet(labels.Set(map[string]string{transformer.Selector: v.(meta.Object).GetName()}))
		options := api.ListOptions{LabelSelector: label}
		komposeLabel := map[string]string{transformer.Selector: v.(meta.Object).GetName()}
		switch t := v.(type) {
		case *extensions.Deployment:
			//delete deployment
			deployment, err := client.Deployments(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range deployment.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpDeployment, err := kubectl.ReaperFor(extensions.Kind("Deployment"), client)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					//FIXME: gracePeriod is nil
					err = rpDeployment.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Deployment: %s", t.Name)

				}
			}

		case *extensions.DaemonSet:
			//delete deployment
			daemonset, err := client.DaemonSets(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range daemonset.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpDaemonset, err := kubectl.ReaperFor(extensions.Kind("DaemonSet"), client)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					//FIXME: gracePeriod is nil
					err = rpDaemonset.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted DaemonSet: %s", t.Name)

				}
			}

		case *api.ReplicationController:
			//delete deployment
			replicationController, err := client.ReplicationControllers(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range replicationController.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpReplicationController, err := kubectl.ReaperFor(api.Kind("ReplicationController"), client)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					//FIXME: gracePeriod is nil
					err = rpReplicationController.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted ReplicationController: %s", t.Name)

				}
			}

		case *api.Service:
			//delete svc
			svc, err := client.Services(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range svc.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpService, err := kubectl.ReaperFor(api.Kind("Service"), client)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					//FIXME: gracePeriod is nil
					err = rpService.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Service: %s", t.Name)

				}
			}

		case *api.PersistentVolumeClaim:
			// delete pvc
			pvc, err := client.PersistentVolumeClaims(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range pvc.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = client.PersistentVolumeClaims(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted PersistentVolumeClaim: %s", t.Name)
				}
			}

		case *extensions.Ingress:
			// delete ingress
			ingDeleteOptions := &api.DeleteOptions{
				TypeMeta: unversioned.TypeMeta{
					Kind:       "Ingress",
					APIVersion: "extensions/v1beta1",
				},
			}
			ingress, err := client.Ingress(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range ingress.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {

					err = client.Ingress(namespace).Delete(t.Name, ingDeleteOptions)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Ingress: %s", t.Name)
				}
			}

		case *api.Pod:
			//delete pod
			pod, err := client.Pods(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
			}
			for _, l := range pod.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpPod, err := kubectl.ReaperFor(api.Kind("Pod"), client)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					//FIXME: gracePeriod is nil
					err = rpPod.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Pod: %s", t.Name)
				}
			}

		case *api.ConfigMap:
			// delete ConfigMap
			configMap, err := client.ConfigMaps(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range configMap.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = client.ConfigMaps(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted ConfigMap: %s", t.Name)
				}
			}
		}
	}

	return errorList
}
