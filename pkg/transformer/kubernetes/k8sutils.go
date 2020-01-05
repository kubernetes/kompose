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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/runtime"

	"sort"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api/resource"
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
	err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"README.md", []byte(readme), 0644)
	if err != nil {
		return err
	}

	/* Create the Chart.yaml file */
	chart := `name: {{.Name}}
description: A generated Helm Chart for {{.Name}} from Skippbox Kompose
version: 0.0.1
apiVersion: v1
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

	err = ioutil.WriteFile(dirName+string(os.PathSeparator)+"Chart.yaml", chartData.Bytes(), 0644)
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
		defer f.Close()
	}

	var files []string

	// if asked to print to stdout or to put in single file
	// we will create a list
	if opt.ToStdout || f != nil {
		list := &api.List{}
		// convert objects to versioned and add them to list
		for _, object := range objects {
			versionedObject, err := convertToVersion(object, unversioned.GroupVersion{})
			if err != nil {
				return err
			}

			list.Items = append(list.Items, versionedObject)

		}
		// version list itself
		listVersion := unversioned.GroupVersion{Group: "", Version: "v1"}
		convertedList, err := convertToVersion(list, listVersion)
		if err != nil {
			return err
		}
		data, err := marshal(convertedList, opt.GenerateJSON, opt.YAMLIndent)
		if err != nil {
			return fmt.Errorf("error in marshalling the List: %v", err)
		}
		printVal, err := transformer.Print("", dirName, "", data, opt.ToStdout, opt.GenerateJSON, f, opt.Provider)
		if err != nil {
			return errors.Wrap(err, "transformer.Print failed")
		}
		files = append(files, printVal)
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
			versionedObject, err := convertToVersion(v, unversioned.GroupVersion{})
			if err != nil {
				return err
			}
			data, err := marshal(versionedObject, opt.GenerateJSON, opt.YAMLIndent)
			if err != nil {
				return err
			}

			var typeMeta unversioned.TypeMeta
			var objectMeta api.ObjectMeta

			if us, ok := v.(*runtime.Unstructured); ok {
				typeMeta = unversioned.TypeMeta{
					Kind:       us.GetKind(),
					APIVersion: us.GetAPIVersion(),
				}
				objectMeta = api.ObjectMeta{
					Name: us.GetName(),
				}
			} else {
				val := reflect.ValueOf(v).Elem()
				// Use reflect to access TypeMeta struct inside runtime.Object.
				// cast it to correct type - unversioned.TypeMeta
				typeMeta = val.FieldByName("TypeMeta").Interface().(unversioned.TypeMeta)

				// Use reflect to access ObjectMeta struct inside runtime.Object.
				// cast it to correct type - api.ObjectMeta
				objectMeta = val.FieldByName("ObjectMeta").Interface().(api.ObjectMeta)

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
// if groupVersion is  empty (unversioned.GroupVersion{}), use version from original object (obj)
func convertToVersion(obj runtime.Object, groupVersion unversioned.GroupVersion) (runtime.Object, error) {

	// ignore unstruct object
	if _, ok := obj.(*runtime.Unstructured); ok {
		return obj, nil
	}

	var version unversioned.GroupVersion

	if groupVersion.Empty() {
		objectVersion := obj.GetObjectKind().GroupVersionKind()
		version = unversioned.GroupVersion{Group: objectVersion.Group, Version: objectVersion.Version}
	} else {
		version = groupVersion
	}
	convertedObject, err := api.Scheme.ConvertToVersion(obj, version)
	if err != nil {
		return nil, err
	}
	return convertedObject, nil
}

// PortsExist checks if service has ports defined
func (k *Kubernetes) PortsExist(service kobject.ServiceConfig) bool {
	return len(service.Port) != 0
}

// CreateService creates a k8s service
func (k *Kubernetes) CreateService(name string, service kobject.ServiceConfig, objects []runtime.Object) *api.Service {
	svc := k.InitSvc(name, service)

	// Configure the service ports.
	servicePorts := k.ConfigServicePorts(name, service)
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
func (k *Kubernetes) CreateHeadlessService(name string, service kobject.ServiceConfig, objects []runtime.Object) *api.Service {
	svc := k.InitSvc(name, service)

	servicePorts := []api.ServicePort{}
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

// UpdateKubernetesObjects loads configurations to k8s objects
func (k *Kubernetes) UpdateKubernetesObjects(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions, objects *[]runtime.Object) error {

	// Configure the environment variables.
	envs, err := k.ConfigEnvs(name, service, opt)
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

	if pvc != nil {
		// Looping on the slice pvc instead of `*objects = append(*objects, pvc...)`
		// because the type of objects and pvc is different, but when doing append
		// one element at a time it gets converted to runtime.Object for objects slice
		for _, p := range pvc {
			*objects = append(*objects, p)
		}
	}

	if cms != nil {
		for _, c := range cms {
			*objects = append(*objects, c)
		}
	}

	// Configure the container ports.
	ports := k.ConfigPorts(name, service)

	// Configure capabilities
	capabilities := k.ConfigCapabilities(service)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)

	// fillTemplate fills the pod template with the value calculated from config
	fillTemplate := func(template *api.PodTemplateSpec) error {
		if len(service.ContainerName) > 0 {
			template.Spec.Containers[0].Name = FormatContainerName(service.ContainerName)
		}
		template.Spec.Containers[0].Env = envs
		template.Spec.Containers[0].Command = service.Command
		template.Spec.Containers[0].Args = service.Args
		template.Spec.Containers[0].WorkingDir = service.WorkingDir
		template.Spec.Containers[0].VolumeMounts = append(template.Spec.Containers[0].VolumeMounts, volumesMount...)
		template.Spec.Containers[0].Stdin = service.Stdin
		template.Spec.Containers[0].TTY = service.Tty
		template.Spec.Volumes = append(template.Spec.Volumes, volumes...)
		template.Spec.NodeSelector = service.Placement
		// Configure the HealthCheck
		// We check to see if it's blank
		if !reflect.DeepEqual(service.HealthChecks, kobject.HealthCheck{}) {
			probe := api.Probe{}

			if len(service.HealthChecks.Test) > 0 {
				probe.Handler = api.Handler{
					Exec: &api.ExecAction{
						Command: service.HealthChecks.Test,
					},
				}
			} else {
				return errors.New("Health check must contain a command")
			}

			probe.TimeoutSeconds = service.HealthChecks.Timeout
			probe.PeriodSeconds = service.HealthChecks.Interval
			probe.FailureThreshold = service.HealthChecks.Retries

			// See issue: https://github.com/docker/cli/issues/116
			// StartPeriod has been added to docker/cli however, it is not yet added
			// to compose. Once the feature has been implemented, this will automatically work
			probe.InitialDelaySeconds = service.HealthChecks.StartPeriod

			template.Spec.Containers[0].LivenessProbe = &probe
		}

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
				podSecurityContext.HostPID = true
			} else {
				log.Warningf("Ignoring PID key for service \"%v\". Invalid value \"%v\".", name, service.Pid)
			}
		}

		//set supplementalGroups
		if service.GroupAdd != nil {
			podSecurityContext.SupplementalGroups = service.GroupAdd
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
		if policy, err := GetImagePullPolicy(name, service.ImagePullPolicy); err != nil {
			return err
		} else {
			template.Spec.Containers[0].ImagePullPolicy = policy
		}

		// Configure the container restart policy.
		if restart, err := GetRestartPolicy(name, service.Restart); err != nil {
			return err
		} else {
			template.Spec.RestartPolicy = restart
		}

		// Configure hostname/domain_name settings
		if service.HostName != "" {
			template.Spec.Hostname = service.HostName
		}
		if service.DomainName != "" {
			template.Spec.Subdomain = service.DomainName
		}

		return nil
	}

	// fillObjectMeta fills the metadata with the value calculated from config
	fillObjectMeta := func(meta *api.ObjectMeta) {
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
			case *extensions.Deployment:
				objType.Spec.Strategy.Type = extensions.RecreateDeploymentStrategyType
			case *deployapi.DeploymentConfig:
				objType.Spec.Strategy.Type = deployapi.DeploymentStrategyTypeRecreate
			}
		}
	}
	return nil
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

	return

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
// since we know for sure that the duplication can only happends on ConfigMap, so
// this code will looks like this for now.
func (k *Kubernetes) RemoveDupObjects(objs *[]runtime.Object) {
	var result []runtime.Object
	exist := map[string]bool{}
	for _, obj := range *objs {
		if us, ok := obj.(*api.ConfigMap); ok {
			k := us.GroupVersionKind().String() + us.GetNamespace() + us.GetName()
			if exist[k] {
				log.Debugf("Remove duplicate configmap: %s", us.GetName())
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

func resetWorkloadAPIVersion(d runtime.Object) runtime.Object {
	data, err := json.Marshal(d)
	if err == nil {
		var us runtime.Unstructured
		if err := json.Unmarshal(data, &us); err == nil {
			us.SetGroupVersionKind(unversioned.GroupVersionKind{
				Group:   "apps",
				Version: "v1",
				Kind:    d.GetObjectKind().GroupVersionKind().Kind,
			})
			return &us
		}
	}
	return d
}

// FixWorkloadVersion force reset deployment/daemonset's apiversion to apps/v1
func (k *Kubernetes) FixWorkloadVersion(objs *[]runtime.Object) {
	var result []runtime.Object
	for _, obj := range *objs {
		if d, ok := obj.(*extensions.Deployment); ok {
			nd := resetWorkloadAPIVersion(d)
			result = append(result, nd)
		} else if d, ok := obj.(*extensions.DaemonSet); ok {
			nd := resetWorkloadAPIVersion(d)
			result = append(result, nd)
		} else {
			result = append(result, obj)
		}
	}
	*objs = result
}

// SortedKeys Ensure the kubernetes objects are in a consistent order
func SortedKeys(komposeObject kobject.KomposeObject) []string {
	var sortedKeys []string
	for name := range komposeObject.ServiceConfigs {
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
func GetEnvsFromFile(file string, opt kobject.ConvertOptions) (map[string]string, error) {
	// Get the correct file context / directory
	composeDir, err := transformer.GetComposeFileDir(opt.InputFiles)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to load file context")
	}
	fileLocation := path.Join(composeDir, file)

	// Load environment variables from file
	envLoad, err := godotenv.Read(fileLocation)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read env_file")
	}

	return envLoad, nil
}

// GetContentFromFile gets the content from the file..
func GetContentFromFile(file string) (string, error) {
	fileBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", errors.Wrap(err, "Unable to read file")
	}
	return string(fileBytes), nil
}

// FormatEnvName format env name
func FormatEnvName(name string) string {
	envName := strings.Trim(name, "./")
	envName = strings.Replace(envName, ".", "-", -1)
	envName = strings.Replace(envName, "/", "-", -1)
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

//FormatContainerName format Container name
func FormatContainerName(name string) string {
	name = strings.Replace(name, "_", "-", -1)
	return name

}
