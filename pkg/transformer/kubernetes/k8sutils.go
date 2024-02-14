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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/compose-spec/compose-go/types"
	"github.com/joho/godotenv"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/loader/compose"
	"github.com/kubernetes/kompose/pkg/transformer"
	deployapi "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	appsv1 "k8s.io/api/apps/v1"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

/**
 * Generate Helm Chart configuration
 */
func generateHelm(dirName string) error {
	type ChartDetails struct {
		Name string
	}

	details := ChartDetails{dirName}
	manifestDir := dirName + string(os.PathSeparator) + "templates"
	dir, err := os.Open(dirName)

	/* Setup the initial directories/files */
	if err == nil {
		_ = dir.Close()
	}

	if err != nil {
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			return err
		}

		err = os.Mkdir(manifestDir, 0755)
		if err != nil {
			return err
		}
	}

	/* Create the readme file */
	readme := "This chart was created by Kompose\n"
	err = os.WriteFile(dirName+string(os.PathSeparator)+"README.md", []byte(readme), 0644)
	if err != nil {
		return err
	}

	/* Create the Chart.yaml file */
	chart := `name: {{.Name}}
description: A generated Helm Chart for {{.Name}} from Skippbox Kompose
version: 0.0.1
apiVersion: v2
keywords:
  - {{.Name}}
sources:
home:
`

	t, err := template.New("ChartTmpl").Parse(chart)
	if err != nil {
		return errors.Wrap(err, "Failed to generate Chart.yaml template, template.New failed")
	}
	var chartData bytes.Buffer
	_ = t.Execute(&chartData, details)

	err = os.WriteFile(dirName+string(os.PathSeparator)+"Chart.yaml", chartData.Bytes(), 0644)
	if err != nil {
		return err
	}

	log.Infof("chart created in %q\n", dirName+string(os.PathSeparator))
	return nil
}

// Check if given path is a directory
func isDir(name string) (bool, error) {
	// Open file to get stat later
	f, err := os.Open(name)
	if err != nil {
		return false, nil
	}
	defer f.Close()

	// Get file attributes and information
	fileStat, err := f.Stat()
	if err != nil {
		return false, errors.Wrap(err, "error retrieving file information, f.Stat failed")
	}

	// Check if given path is a directory
	if fileStat.IsDir() {
		return true, nil
	}
	return false, nil
}

func getDirName(opt kobject.ConvertOptions) string {
	dirName := opt.OutFile
	if dirName == "" {
		// Let assume all the docker-compose files are in the same directory
		if opt.CreateChart {
			filename := opt.InputFiles[0]
			extension := filepath.Ext(filename)
			dirName = filename[0 : len(filename)-len(extension)]
		} else {
			dirName = "."
		}
	}
	return dirName
}

// PrintList will take the data converted and decide on the commandline attributes given
func PrintList(objects []runtime.Object, opt kobject.ConvertOptions) error {
	var f *os.File
	dirName := getDirName(opt)
	log.Debugf("Target Dir: %s", dirName)

	// Create a directory if "out" ends with "/" and does not exist.
	if !transformer.Exists(opt.OutFile) && strings.HasSuffix(opt.OutFile, "/") {
		if err := os.MkdirAll(opt.OutFile, os.ModePerm); err != nil {
			return errors.Wrap(err, "failed to create a directory")
		}
	}

	// Check if output file is a directory
	isDirVal, err := isDir(opt.OutFile)
	if err != nil {
		return errors.Wrap(err, "isDir failed")
	}
	if opt.CreateChart {
		isDirVal = true
	}
	if !isDirVal {
		f, err = transformer.CreateOutFile(opt.OutFile)
		if err != nil {
			return errors.Wrap(err, "transformer.CreateOutFile failed")
		}
		if len(opt.OutFile) != 0 {
			log.Printf("Kubernetes file %q created", opt.OutFile)
		}
		defer f.Close()
	}

	var files []string
	// if asked to print to stdout or to put in single file
	// we will create a list
	if opt.ToStdout || f != nil {
		// convert objects to versioned and add them to list
		if opt.GenerateJSON {
			return fmt.Errorf("cannot convert to one file while specifying a json output file or stdout option")
		}
		for _, object := range objects {
			versionedObject, err := convertToVersion(object)
			if err != nil {
				return err
			}

			data, err := marshal(versionedObject, opt.GenerateJSON, opt.YAMLIndent)
			if err != nil {
				return fmt.Errorf("error in marshalling the List: %v", err)
			}
			// this part add --- which unifies the file
			data = []byte(fmt.Sprintf("---\n%s", data))
			printVal, err := transformer.Print("", dirName, "", data, opt.ToStdout, opt.GenerateJSON, f, opt.Provider)
			if err != nil {
				return errors.Wrap(err, "transformer to print to one single file failed")
			}
			files = append(files, printVal)
		}
	} else {
		finalDirName := dirName
		if opt.CreateChart {
			finalDirName = dirName + string(os.PathSeparator) + "templates"
		}

		if err := os.MkdirAll(finalDirName, 0755); err != nil {
			return err
		}

		var file string
		// create a separate file for each provider
		for _, v := range objects {
			versionedObject, err := convertToVersion(v)
			if err != nil {
				return err
			}
			data, err := marshal(versionedObject, opt.GenerateJSON, opt.YAMLIndent)
			if err != nil {
				return err
			}

			var typeMeta metav1.TypeMeta
			var objectMeta metav1.ObjectMeta

			if us, ok := v.(*unstructured.Unstructured); ok {
				typeMeta = metav1.TypeMeta{
					Kind:       us.GetKind(),
					APIVersion: us.GetAPIVersion(),
				}
				objectMeta = metav1.ObjectMeta{
					Name: us.GetName(),
				}
			} else {
				val := reflect.ValueOf(v).Elem()
				// Use reflect to access TypeMeta struct inside runtime.Object.
				// cast it to correct type - metav1.TypeMeta
				typeMeta = val.FieldByName("TypeMeta").Interface().(metav1.TypeMeta)

				// Use reflect to access ObjectMeta struct inside runtime.Object.
				// cast it to correct type - api.ObjectMeta
				objectMeta = val.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
			}

			file, err = transformer.Print(objectMeta.Name, finalDirName, strings.ToLower(typeMeta.Kind), data, opt.ToStdout, opt.GenerateJSON, f, opt.Provider)
			if err != nil {
				return errors.Wrap(err, "transformer.Print failed")
			}

			files = append(files, file)
		}
	}
	if opt.CreateChart {
		err = generateHelm(dirName)
		if err != nil {
			return errors.Wrap(err, "generateHelm failed")
		}
	}
	return nil
}

// marshal object runtime.Object and return byte array
func marshal(obj runtime.Object, jsonFormat bool, indent int) (data []byte, err error) {
	// convert data to yaml or json
	if jsonFormat {
		data, err = json.MarshalIndent(obj, "", "  ")
	} else {
		data, err = marshalWithIndent(obj, indent)
	}
	if err != nil {
		data = nil
	}
	return
}

// remove empty map[string]interface{} strings from the object
//
// Note: this function uses recursion, use it only objects created by the unmarshalled json.
// Passing cyclic structures to removeEmptyInterfaces will result in a stack overflow.
func removeEmptyInterfaces(obj interface{}) interface{} {
	switch v := obj.(type) {
	case []interface{}:
		for i, val := range v {
			if valMap, ok := val.(map[string]interface{}); (ok && len(valMap) == 0) || val == nil {
				v = append(v[:i], v[i+1:]...)
			} else {
				v[i] = removeEmptyInterfaces(val)
			}
		}
		return v
	case map[string]interface{}:
		for k, val := range v {
			if valMap, ok := val.(map[string]interface{}); ok {
				// It is always map[string]interface{} when passed the map[string]interface{}
				valMap := removeEmptyInterfaces(valMap).(map[string]interface{})
				if len(valMap) == 0 {
					delete(v, k)
				}
			} else if val == nil {
				delete(v, k)
			} else {
				processedInterface := removeEmptyInterfaces(val)
				if valSlice, ok := processedInterface.([]interface{}); ok && len(valSlice) == 0 {
					delete(v, k)
				} else {
					v[k] = processedInterface
				}
			}
		}
		return v
	default:
		return v
	}
}

// Convert JSON to YAML.
func jsonToYaml(j []byte, spaces int) ([]byte, error) {
	// Convert the JSON to an object.
	var jsonObj interface{}
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	err := yaml.Unmarshal(j, &jsonObj)
	if err != nil {
		return nil, err
	}
	jsonObj = removeEmptyInterfaces(jsonObj)
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(spaces)
	if err := encoder.Encode(jsonObj); err != nil {
		return nil, err
	}
	return b.Bytes(), nil

	// Marshal this object into YAML.
	// return yaml.Marshal(jsonObj)
}

func marshalWithIndent(o interface{}, indent int) ([]byte, error) {
	j, err := json.Marshal(o)
	if err != nil {
		return nil, fmt.Errorf("error marshaling into JSON: %s", err.Error())
	}

	y, err := jsonToYaml(j, indent)
	if err != nil {
		return nil, fmt.Errorf("error converting JSON to YAML: %s", err.Error())
	}

	return y, nil
}

// Convert object to versioned object
// if groupVersion is  empty (metav1.GroupVersion{}), use version from original object (obj)
func convertToVersion(obj runtime.Object) (runtime.Object, error) {
	// ignore unstruct object
	if _, ok := obj.(*unstructured.Unstructured); ok {
		return obj, nil
	}

	return obj, nil

	//var version metav1.GroupVersion
	//
	//if groupVersion.Empty() {
	//	objectVersion := obj.GetObjectKind().GroupVersionKind()
	//	version = metav1.GroupVersion{Group: objectVersion.Group, Version: objectVersion.Version}
	//} else {
	//	version = groupVersion
	//}
	//convertedObject, err := api.Scheme.ConvertToVersion(obj, version)
	//if err != nil {
	//	return nil, err
	//}
	//return convertedObject, nil
}

// PortsExist checks if service has ports defined
func (k *Kubernetes) PortsExist(service kobject.ServiceConfig) bool {
	return len(service.Port) != 0
}

func (k *Kubernetes) initSvcObject(name string, service kobject.ServiceConfig, ports []api.ServicePort) *api.Service {
	svc := k.InitSvc(name, service)
	// special case, only for loaderbalancer type
	svc.Name = name
	svc.Spec.Selector = transformer.ConfigLabels(service.Name)

	svc.Spec.Ports = ports
	svc.Spec.Type = api.ServiceType(service.ServiceType)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// CreateLBService creates a k8s Load Balancer Service
func (k *Kubernetes) CreateLBService(name string, service kobject.ServiceConfig) []*api.Service {
	var svcs []*api.Service
	tcpPorts, udpPorts := k.ConfigLBServicePorts(service)
	if tcpPorts != nil {
		svc := k.initSvcObject(name+"-tcp", service, tcpPorts)
		svcs = append(svcs, svc)
	}
	if udpPorts != nil {
		svc := k.initSvcObject(name+"-udp", service, udpPorts)
		svcs = append(svcs, svc)
	}
	return svcs
}

// CreateService creates a k8s service
func (k *Kubernetes) CreateService(name string, service kobject.ServiceConfig) *api.Service {
	svc := k.InitSvc(name, service)

	// Configure the service ports.
	servicePorts := k.ConfigServicePorts(service)
	svc.Spec.Ports = servicePorts

	if service.ServiceType == "Headless" {
		svc.Spec.Type = api.ServiceTypeClusterIP
		svc.Spec.ClusterIP = "None"
	} else {
		svc.Spec.Type = api.ServiceType(service.ServiceType)
	}

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// CreateHeadlessService creates a k8s headless service.
// This is used for docker-compose services without ports. For such services we can't create regular Kubernetes Service.
// and without Service Pods can't find each other using DNS names.
// Instead of regular Kubernetes Service we create Headless Service. DNS of such service points directly to Pod IP address.
// You can find more about Headless Services in Kubernetes documentation https://kubernetes.io/docs/user-guide/services/#headless-services
func (k *Kubernetes) CreateHeadlessService(name string, service kobject.ServiceConfig) *api.Service {
	svc := k.InitSvc(name, service)

	var servicePorts []api.ServicePort
	// Configure a dummy port: https://github.com/kubernetes/kubernetes/issues/32766.
	servicePorts = append(servicePorts, api.ServicePort{
		Name: "headless",
		Port: 55555,
	})

	svc.Spec.Ports = servicePorts
	svc.Spec.ClusterIP = "None"

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// UpdateKubernetesObjectsMultipleContainers method updates the kubernetes objects with the necessary data
func (k *Kubernetes) UpdateKubernetesObjectsMultipleContainers(name string, service kobject.ServiceConfig, objects *[]runtime.Object, podSpec PodSpec) error {
	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)

	// fillTemplate fills the pod template with the value calculated from config
	fillTemplate := func(template *api.PodTemplateSpec) error {
		template.ObjectMeta.Labels = transformer.ConfigLabelsWithNetwork(name, service.Network)
		template.Spec = podSpec.Get()
		return nil
	}

	// fillObjectMeta fills the metadata with the value calculated from config
	fillObjectMeta := func(meta *metav1.ObjectMeta) {
		meta.Annotations = annotations
	}

	// update supported controller
	for _, obj := range *objects {
		err := k.UpdateController(obj, fillTemplate, fillObjectMeta)
		if err != nil {
			return errors.Wrap(err, "k.UpdateController failed")
		}
		if len(service.Volumes) > 0 {
			switch objType := obj.(type) {
			case *appsv1.Deployment:
				objType.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
			case *deployapi.DeploymentConfig:
				objType.Spec.Strategy.Type = deployapi.DeploymentStrategyTypeRecreate
			}
		}
	}
	return nil
}

// UpdateKubernetesObjects loads configurations to k8s objects
func (k *Kubernetes) UpdateKubernetesObjects(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions, objects *[]runtime.Object) error {
	// Configure the environment variables.
	envs, err := ConfigEnvs(service, opt)
	if err != nil {
		return errors.Wrap(err, "Unable to load env variables")
	}

	// Configure the container volumes.
	volumesMount, volumes, pvc, cms, err := k.ConfigVolumes(name, service)
	if err != nil {
		return errors.Wrap(err, "k.ConfigVolumes failed")
	}
	// Configure Tmpfs
	if len(service.TmpFs) > 0 {
		TmpVolumesMount, TmpVolumes := k.ConfigTmpfs(name, service)
		volumes = append(volumes, TmpVolumes...)
		volumesMount = append(volumesMount, TmpVolumesMount...)
	}

	if pvc != nil && opt.Controller != StatefulStateController {
		// Looping on the slice pvc instead of `*objects = append(*objects, pvc...)`
		// because the type of objects and pvc is different, but when doing append
		// one element at a time it gets converted to runtime.Object for objects slice
		for _, p := range pvc {
			*objects = append(*objects, p)
		}
	}

	for _, c := range cms {
		*objects = append(*objects, c)
	}

	// Configure the container ports.
	ports := ConfigPorts(service)
	// Configure capabilities
	capabilities := ConfigCapabilities(service)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)

	// fillTemplate fills the pod template with the value calculated from config
	fillTemplate := func(template *api.PodTemplateSpec) error {
		template.Spec.Containers[0].Name = GetContainerName(service)
		template.Spec.Containers[0].Env = envs
		template.Spec.Containers[0].Command = service.Command
		template.Spec.Containers[0].Args = GetContainerArgs(service)
		template.Spec.Containers[0].WorkingDir = service.WorkingDir
		template.Spec.Containers[0].VolumeMounts = append(template.Spec.Containers[0].VolumeMounts, volumesMount...)
		template.Spec.Containers[0].Stdin = service.Stdin
		template.Spec.Containers[0].TTY = service.Tty
		if opt.Controller != StatefulStateController || opt.Volumes == "configMap" {
			template.Spec.Volumes = append(template.Spec.Volumes, volumes...)
		}
		template.Spec.Affinity = ConfigAffinity(service)
		template.Spec.TopologySpreadConstraints = ConfigTopologySpreadConstraints(service)
		// Configure the HealthCheck
		template.Spec.Containers[0].LivenessProbe = configProbe(service.HealthChecks.Liveness)
		template.Spec.Containers[0].ReadinessProbe = configProbe(service.HealthChecks.Readiness)

		if service.StopGracePeriod != "" {
			template.Spec.TerminationGracePeriodSeconds, err = DurationStrToSecondsInt(service.StopGracePeriod)
			if err != nil {
				log.Warningf("Failed to parse duration \"%v\" for service \"%v\"", service.StopGracePeriod, name)
			}
		}

		TranslatePodResource(&service, template)

		// Configure resource reservations
		podSecurityContext := &api.PodSecurityContext{}

		//set pid namespace mode
		if service.Pid != "" {
			if service.Pid == "host" {
				// podSecurityContext.HostPID = true
			} else {
				log.Warningf("Ignoring PID key for service \"%v\". Invalid value \"%v\".", name, service.Pid)
			}
		}

		//set supplementalGroups
		if service.GroupAdd != nil {
			podSecurityContext.SupplementalGroups = service.GroupAdd
		}

		//set Security Context FsGroup
		if service.FsGroup != 0 {
			podSecurityContext.FSGroup = &service.FsGroup
		}

		// Setup security context
		securityContext := &api.SecurityContext{}
		if service.Privileged {
			securityContext.Privileged = &service.Privileged
		}
		if service.User != "" {
			uid, err := strconv.ParseInt(service.User, 10, 64)
			if err != nil {
				log.Warn("Ignoring user directive. User to be specified as a UID (numeric).")
			} else {
				securityContext.RunAsUser = &uid
			}
		}

		//set capabilities if it is not empty
		if len(capabilities.Add) > 0 || len(capabilities.Drop) > 0 {
			securityContext.Capabilities = capabilities
		}

		//set readOnlyRootFilesystem if it is enabled
		if service.ReadOnly {
			securityContext.ReadOnlyRootFilesystem = &service.ReadOnly
		}

		// update template only if securityContext is not empty
		if *securityContext != (api.SecurityContext{}) {
			template.Spec.Containers[0].SecurityContext = securityContext
		}
		if !reflect.DeepEqual(*podSecurityContext, api.PodSecurityContext{}) {
			template.Spec.SecurityContext = podSecurityContext
		}
		template.Spec.Containers[0].Ports = ports
		template.ObjectMeta.Labels = transformer.ConfigLabelsWithNetwork(name, service.Network)

		// Configure the image pull policy
		policy, err := GetImagePullPolicy(name, service.ImagePullPolicy)
		if err != nil {
			return err
		}
		template.Spec.Containers[0].ImagePullPolicy = policy

		// Configure the container restart policy.
		restart, err := GetRestartPolicy(name, service.Restart)
		if err != nil {
			return err
		}
		template.Spec.RestartPolicy = restart

		// Configure hostname/domain_name settings
		if service.HostName != "" {
			template.Spec.Hostname = service.HostName
		}
		if service.DomainName != "" {
			template.Spec.Subdomain = service.DomainName
		}

		if serviceAccountName, ok := service.Labels[compose.LabelServiceAccountName]; ok {
			template.Spec.ServiceAccountName = serviceAccountName
		}

		return nil
	}

	// fillObjectMeta fills the metadata with the value calculated from config
	fillObjectMeta := func(meta *metav1.ObjectMeta) {
		meta.Annotations = annotations
	}

	// update supported controller
	for _, obj := range *objects {
		err = k.UpdateController(obj, fillTemplate, fillObjectMeta)
		if err != nil {
			return errors.Wrap(err, "k.UpdateController failed")
		}
		if len(service.Volumes) > 0 {
			switch objType := obj.(type) {
			case *appsv1.Deployment:
				objType.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
			case *deployapi.DeploymentConfig:
				objType.Spec.Strategy.Type = deployapi.DeploymentStrategyTypeRecreate
			case *appsv1.StatefulSet:
				// embed all PVCs inside the StatefulSet object
				if opt.Volumes == "configMap" {
					break
				}
				persistentVolumeClaims := make([]api.PersistentVolumeClaim, len(pvc))
				for i, persistentVolumeClaim := range pvc {
					persistentVolumeClaims[i] = *persistentVolumeClaim
					persistentVolumeClaims[i].APIVersion = ""
					persistentVolumeClaims[i].Kind = ""
				}
				objType.Spec.VolumeClaimTemplates = persistentVolumeClaims
			}
		}
	}
	return nil
}

// getServiceVolumesID create a unique id for the service's volume mounts
func getServiceVolumesID(service kobject.ServiceConfig) string {
	id := ""
	for _, v := range service.VolList {
		id += v
	}
	return id
}

// getServiceGroupID ...
// return empty string should mean this service should go alone
func getServiceGroupID(service kobject.ServiceConfig, mode string) string {
	if mode == "label" {
		return service.Labels[compose.LabelServiceGroup]
	}
	if mode == "volume" {
		return getServiceVolumesID(service)
	}
	return ""
}

// KomposeObjectToServiceConfigGroupMapping returns the service config group by name or by volume
// This group function works as following
//  1. Support two mode
//     (1): label: use a custom label, the service that contains it will be merged to one workload.
//     (2): volume: the service that share to exactly same volume config will be merged to one workload. If use pvc, only
//     create one for this group.
//  2. If service containers restart policy and no workload argument provide and it's restart policy looks like a pod, then
//     this service should generate a pod. If group mode specified, it should be grouped and ignore the restart policy.
//  3. If group mode specified, port conflict between services in one group will be ignored, and multiple service should be created.
//  4. If `volume` group mode specified, we don't have an appropriate name for this combined service, use the first one for now.
//     A warn/info message should be printed to let the user know.
func KomposeObjectToServiceConfigGroupMapping(komposeObject *kobject.KomposeObject, opt kobject.ConvertOptions) map[string]kobject.ServiceConfigGroup {
	serviceConfigGroup := make(map[string]kobject.ServiceConfigGroup)
	sortedServiceConfigs := SortedKeys(komposeObject.ServiceConfigs)

	for _, service := range sortedServiceConfigs {
		serviceConfig := komposeObject.ServiceConfigs[service]
		groupID := getServiceGroupID(serviceConfig, opt.ServiceGroupMode)
		if groupID != "" {
			serviceConfig.Name = service
			serviceConfig.InGroup = true
			serviceConfigGroup[groupID] = append(serviceConfigGroup[groupID], serviceConfig)
			komposeObject.ServiceConfigs[service] = serviceConfig
		}
	}

	return serviceConfigGroup
}

// TranslatePodResource config pod resources
func TranslatePodResource(service *kobject.ServiceConfig, template *api.PodTemplateSpec) {
	// Configure the resource limits
	if service.MemLimit != 0 || service.CPULimit != 0 {
		resourceLimit := api.ResourceList{}

		if service.MemLimit != 0 {
			resourceLimit[api.ResourceMemory] = *resource.NewQuantity(int64(service.MemLimit), "RandomStringForFormat")
		}

		if service.CPULimit != 0 {
			resourceLimit[api.ResourceCPU] = *resource.NewMilliQuantity(service.CPULimit, resource.DecimalSI)
		}

		template.Spec.Containers[0].Resources.Limits = resourceLimit
	}

	// Configure the resource requests
	if service.MemReservation != 0 || service.CPUReservation != 0 {
		resourceRequests := api.ResourceList{}

		if service.MemReservation != 0 {
			resourceRequests[api.ResourceMemory] = *resource.NewQuantity(int64(service.MemReservation), "RandomStringForFormat")
		}

		if service.CPUReservation != 0 {
			resourceRequests[api.ResourceCPU] = *resource.NewMilliQuantity(service.CPUReservation, resource.DecimalSI)
		}

		template.Spec.Containers[0].Resources.Requests = resourceRequests
	}
}

// GetImagePullPolicy get image pull settings
func GetImagePullPolicy(name, policy string) (api.PullPolicy, error) {
	switch policy {
	case "":
	case "Always":
		return api.PullAlways, nil
	case "Never":
		return api.PullNever, nil
	case "IfNotPresent":
		return api.PullIfNotPresent, nil
	default:
		return "", errors.New("Unknown image-pull-policy " + policy + " for service " + name)
	}
	return "", nil
}

// GetRestartPolicy ...
func GetRestartPolicy(name, restart string) (api.RestartPolicy, error) {
	switch restart {
	case "", "always", "any":
		return api.RestartPolicyAlways, nil
	case "no", "none":
		return api.RestartPolicyNever, nil
	case "on-failure":
		return api.RestartPolicyOnFailure, nil
	default:
		return "", errors.New("Unknown restart policy " + restart + " for service " + name)
	}
}

// SortServicesFirst - the objects that we get can be in any order this keeps services first
// according to best practice kubernetes services should be created first
// http://kubernetes.io/docs/user-guide/config-best-practices/
func (k *Kubernetes) SortServicesFirst(objs *[]runtime.Object) {
	var svc, others, ret []runtime.Object

	for _, obj := range *objs {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Service" {
			svc = append(svc, obj)
		} else {
			others = append(others, obj)
		}
	}
	ret = append(ret, svc...)
	ret = append(ret, others...)
	*objs = ret
}

// RemoveDupObjects remove objects that are dups...eg. configmaps from env.
// since we know for sure that the duplication can only happen on ConfigMap, so
// this code will looks like this for now.
// + NetworkPolicy
func (k *Kubernetes) RemoveDupObjects(objs *[]runtime.Object) {
	var result []runtime.Object
	exist := map[string]bool{}
	for _, obj := range *objs {
		if us, ok := obj.(metav1.Object); ok {
			k := obj.GetObjectKind().GroupVersionKind().String() + us.GetNamespace() + us.GetName()
			if exist[k] {
				log.Debugf("Remove duplicate resource: %s/%s", obj.GetObjectKind().GroupVersionKind().Kind, us.GetName())
				continue
			} else {
				result = append(result, obj)
				exist[k] = true
			}
		} else {
			result = append(result, obj)
		}
	}
	*objs = result
}

// SortedKeys Ensure the kubernetes objects are in a consistent order
func SortedKeys[V kobject.ServiceConfig | kobject.ServiceConfigGroup](serviceConfig map[string]V) []string {
	var sortedKeys []string
	for name := range serviceConfig {
		sortedKeys = append(sortedKeys, name)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

// DurationStrToSecondsInt converts duration string to *int64 in seconds
func DurationStrToSecondsInt(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}
	r := (int64)(duration.Seconds())
	return &r, nil
}

// GetEnvsFromFile get env vars from env_file
func GetEnvsFromFile(file string) (map[string]string, error) {

	envLoad, err := godotenv.Read(file)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read env_file")
	}

	return envLoad, nil
}

// GetContentFromFile gets the content from the file..
func GetContentFromFile(file string) (string, error) {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return "", errors.Wrap(err, "Unable to read file")
	}
	return string(fileBytes), nil
}

// FormatEnvName format env name
func FormatEnvName(name string, serviceName string) string {
	envName := strings.Trim(name, "./")
	// only take string after the last slash only if the string contains a slash
	if strings.Contains(envName, "/") {
		envName = envName[strings.LastIndex(envName, "/")+1:]
	}

	envName = strings.Replace(envName, ".", "-", -1)
	envName = getUsableNameEnvFile(envName, serviceName)
	return envName
}

// getUsableNameEnvFile checks and adjusts the environment file name to make it usable.
// If the first character of envName is a hyphen "-", it is concatenated with nameService.
// If the length of envName is greater than 63, it is truncated to 63 characters.
// Returns the adjusted environment file name.
func getUsableNameEnvFile(envName string, serviceName string) string {
	if string(envName[0]) == "-" { // -env-local....
		envName = fmt.Sprintf("%s%s", serviceName, envName)
	}
	if len(envName) > 63 {
		envName = envName[0:63]
	}
	return envName
}

// FormatFileName format file name
func FormatFileName(name string) string {
	// Split the filepath name so that we use the
	// file name (after the base) for ConfigMap,
	// it shouldn't matter whether it has special characters or not
	_, file := path.Split(name)

	// Make it DNS-1123 compliant for Kubernetes
	return strings.Replace(file, "_", "-", -1)
}

// FormatContainerName format Container name
func FormatContainerName(name string) string {
	name = strings.Replace(name, "_", "-", -1)
	return name
}

// GetContainerName returns the name of the container, from the service config object
func GetContainerName(service kobject.ServiceConfig) string {
	name := service.Name
	if len(service.ContainerName) > 0 {
		name = service.ContainerName
	}
	return FormatContainerName(name)
}

// FormatResourceName generate a valid k8s resource name
func FormatResourceName(name string) string {
	return strings.ToLower(strings.Replace(name, "_", "-", -1))
}

// GetContainerArgs update the interpolation of env variables if exists.
// example: [curl, $PROTOCOL://$DOMAIN] => [curl, $(PROTOCOL)://$(DOMAIN)]
func GetContainerArgs(service kobject.ServiceConfig) []string {
	var args []string
	re := regexp.MustCompile(`\$([a-zA-Z0-9]*)`)
	for _, arg := range service.Args {
		arg = re.ReplaceAllString(arg, `$($1)`)
		args = append(args, arg)
	}
	return args
}

// GetFileName extracts the file name from a given file path or file name.
// If the input fileName contains a "/", it retrieves the substring after the last "/".
// The function does not format the file name further, as it may contain periods or other valid characters.
// Returns the extracted file name.
func GetFileName(fileName string) string {
	if strings.Contains(fileName, "/") {
		fileName = fileName[strings.LastIndex(fileName, "/")+1:]
	}
	// Not format filename because can begin with .fileName
	return fileName
}

// reformatSecretConfigUnderscoreWithDash takes a ServiceSecretConfig object as input and returns a new instance of ServiceSecretConfig
// where the values of Source and Target are formatted using the FormatResourceName function to replace underscores with dashes and lowercase,
// while the other fields remain unchanged. This is done to ensure consistency in the format of container names within the service's secret configuration.
// this function ensures that source, target names are in an acceptable format for Kubernetes and other systems that may require a specific naming format.
func reformatSecretConfigUnderscoreWithDash(secretConfig types.ServiceSecretConfig) types.ServiceSecretConfig {
	newSecretConfig := types.ServiceSecretConfig{
		Source:     FormatResourceName(secretConfig.Source),
		Target:     FormatResourceName(secretConfig.Target),
		UID:        secretConfig.UID,
		GID:        secretConfig.GID,
		Mode:       secretConfig.Mode,
		Extensions: secretConfig.Extensions,
	}

	return newSecretConfig
}
