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

	"github.com/kubernetes/kompose/pkg/loader/compose"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/labels"
	"path/filepath"
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

	if len(service.Configs) > 0 && service.Configs[0].Mode != nil {
		//This is for LONG SYNTAX
		for _, value := range service.Configs {
			if value.Target == "/" {
				log.Warnf("Long syntax config, target path can not be /")
				continue
			}
			tmpKey := FormatFileName(value.Source)
			volumeMounts = append(volumeMounts,
				api.VolumeMount{
					Name:      tmpKey,
					MountPath: "/" + FormatFileName(value.Target),
				})

			tmpVolume := api.Volume{
				Name: tmpKey,
			}
			tmpVolume.ConfigMap = &api.ConfigMapVolumeSource{}
			tmpVolume.ConfigMap.Name = tmpKey
			var tmpMode int32
			tmpMode = int32(*value.Mode)
			tmpVolume.ConfigMap.DefaultMode = &tmpMode
			volumes = append(volumes, tmpVolume)
		}
	} else {
		//This is for SHORT SYNTAX, unsupported
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
			Name:   name + "-" + envName,
			Labels: transformer.ConfigLabels(name + "-" + envName),
		},
		Data: envs,
	}

	return configMap
}

//InitConfigMapFromFile initializes a ConfigMap object
func (k *Kubernetes) InitConfigMapFromFile(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions, fileName string) *api.ConfigMap {
	content, err := GetContentFromFile(fileName, opt)
	if err != nil {
		log.Fatalf("Unable to retrieve file: %s", err)
	}

	originFileName := FormatFileName(fileName)
	dataMap := make(map[string]string)
	dataMap[originFileName] = content
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
			Labels: transformer.ConfigLabels(name),
		},
		Spec: extensions.DeploymentSpec{
			Replicas: int32(replicas),
			Template: api.PodTemplateSpec{
				Spec: podSpec,
			},
		},
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
			Labels: transformer.ConfigLabels(name),
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
		ingress.Spec.Rules[i] = extensions.IngressRule{
			IngressRuleValue: extensions.IngressRuleValue{
				HTTP: &extensions.HTTPIngressRuleValue{
					Paths: []extensions.HTTPIngressPath{
						{
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
	for _, port := range service.Port {

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

// ConfigVolumes configure the container volumes.
func (k *Kubernetes) ConfigVolumes(name string, service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume, []*api.PersistentVolumeClaim, error) {
	volumeMounts := []api.VolumeMount{}
	volumes := []api.Volume{}
	var PVCs []*api.PersistentVolumeClaim
	var volumeName string

	// Set a var based on if the user wants to use empty volumes
	// as opposed to persistent volumes and volume claims
	useEmptyVolumes := k.Opt.EmptyVols
	useHostPath := false

	if k.Opt.Volumes == "emptyDir" {
		useEmptyVolumes = true
	}

	if k.Opt.Volumes == "hostPath" {
		useHostPath = true
	}

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
		volumeMounts = append(volumeMounts, volMount)
		// Get a volume source based on the type of volume we are using
		// For PVC we will also create a PVC object and add to list
		var volsource *api.VolumeSource

		if useEmptyVolumes {
			volsource = k.ConfigEmptyVolumeSource("volume")
		} else if useHostPath {
			source, err := k.ConfigHostPathVolumeSource(volume.Host)
			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "k.ConfigHostPathVolumeSource failed")
			}
			volsource = source
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
					return nil, nil, nil, errors.Wrap(err, "k.CreatePVC failed")
				}

				PVCs = append(PVCs, createdPVC)
			}

		}

		// create a new volume object using the volsource and add to list
		vol := api.Volume{
			Name:         volumeName,
			VolumeSource: *volsource,
		}
		volumes = append(volumes, vol)

		if len(volume.Host) > 0 && !useHostPath {
			log.Warningf("Volume mount on the host %q isn't supported - ignoring path on the host", volume.Host)
		}

	}

	return volumeMounts, volumes, PVCs, nil
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

// ConfigHostPathVolumeSource is a helper function to create a HostPath api.VolumeSource
func (k *Kubernetes) ConfigHostPathVolumeSource(path string) (*api.VolumeSource, error) {
	dir, err := transformer.GetComposeFileDir(k.Opt.InputFiles)
	version, err := transformer.GetVersionFromFile(k.Opt.InputFiles)
	if err != nil {
		return nil, err
	}
	// Concat path based on version
	switch version {
	// Concat dir with path if it's 1 or 2
	// If blank, it's assumed it's 1 or 2
	case "", "1", "1.0", "2", "2.0":
		return &api.VolumeSource{
			HostPath: &api.HostPathVolumeSource{Path: filepath.Join(dir, path)},
		}, nil
	// Again, in v3, we use the "long syntax" for volumes in terms of parsing
	// https://docs.docker.com/compose/compose-file/#long-syntax-3
	// So the path is already an absolute path
	case "3", "3.0", "3.1", "3.2", "3.3":
		return &api.VolumeSource{
			HostPath: &api.HostPathVolumeSource{Path: path},
		}, nil
	default:
		return &api.VolumeSource{}, fmt.Errorf("Version %s of Docker Compose is not supported. Please use version 1, 2 or 3", version)
	}
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
								Name: name + "-" + envName,
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
		configMap := k.InitConfigMapFromFile(name, service, opt, currentFileName)
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

// Transform maps komposeObject to k8s objects
// returns object that are already sorted in the way that Services are first
func (k *Kubernetes) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) ([]runtime.Object, error) {

	// this will hold all the converted data
	var allobjects []runtime.Object

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

			if service.Image == "" {
				return nil, fmt.Errorf("image key required within build parameters in order to build and push service '%s'", name)
			}

			log.Infof("Build key detected. Attempting to build and push image '%s'", service.Image)

			// Get the directory where the compose file is
			composeFileDir, err := transformer.GetComposeFileDir(opt.InputFiles)
			if err != nil {
				return nil, err
			}

			// Build the image!
			err = transformer.BuildDockerImage(service, name, composeFileDir)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to build Docker image for service %v", name)
			}

			// Push the built image to the repo!
			err = transformer.PushDockerImage(service, name)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to push Docker image for service %v", name)
			}

		}

		// If there's no "image" key, use the name of the container that's built
		if service.Image == "" {
			service.Image = name
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

		allobjects = append(allobjects, objects...)
	}

	// sort all object so Services are first
	k.SortServicesFirst(&allobjects)
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

// GetKubernetesClient creates the k8s Client, returns k8s client and namespace
func (k *Kubernetes) GetKubernetesClient() (*client.Client, string, error) {
	// initialize Kubernetes client
	factory := cmdutil.NewFactory(nil)
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
