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

package kubernetes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer"

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
func generateHelm(filenames []string, outFiles []string) error {
	type ChartDetails struct {
		Name string
	}
	// Let assume all the docker-compose files are in the same directory
	filename := filenames[0]
	extension := filepath.Ext(filename)
	dirName := filename[0 : len(filename)-len(extension)]
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
	}

	/* Copy all related json/yaml files into the newly created manifests directory */
	for _, filename := range outFiles {
		if err = cpFileToChart(manifestDir, filename); err != nil {
			log.Warningln(err)
		}
		if err = os.Remove(filename); err != nil {
			log.Warningln(err)
		}
	}
	log.Infof("chart created in %q\n", "."+string(os.PathSeparator)+dirName+string(os.PathSeparator))
	return nil
}

func cpFileToChart(manifestDir, filename string) error {
	infile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Warningf("Error reading %s: %s\n", filename, err)
		return err
	}

	return ioutil.WriteFile(manifestDir+string(os.PathSeparator)+filename, infile, 0644)
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

// PrintList will take the data converted and decide on the commandline attributes given
func PrintList(objects []runtime.Object, opt kobject.ConvertOptions) error {

	var f *os.File
	var dirName string

	// Check if output file is a directory
	isDirVal, err := isDir(opt.OutFile)
	if err != nil {
		return errors.Wrap(err, "isDir failed")
	}
	if isDirVal {
		dirName = opt.OutFile
	} else {
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
		data, err := marshal(convertedList, opt.GenerateJSON)
		if err != nil {
			return fmt.Errorf("Error in marshalling the List: %v", err)
		}
		printVal, err := transformer.Print("", dirName, "", data, opt.ToStdout, opt.GenerateJSON, f)
		if err != nil {
			return errors.Wrap(err, "transformer.Print failed")
		}
		files = append(files, printVal)
	} else {
		var file string
		// create a separate file for each provider
		for _, v := range objects {
			versionedObject, err := convertToVersion(v, unversioned.GroupVersion{})
			if err != nil {
				return err
			}
			data, err := marshal(versionedObject, opt.GenerateJSON)
			if err != nil {
				return err
			}

			val := reflect.ValueOf(v).Elem()
			// Use reflect to access TypeMeta struct inside runtime.Object.
			// cast it to correct type - unversioned.TypeMeta
			typeMeta := val.FieldByName("TypeMeta").Interface().(unversioned.TypeMeta)

			// Use reflect to access ObjectMeta struct inside runtime.Object.
			// cast it to correct type - api.ObjectMeta
			objectMeta := val.FieldByName("ObjectMeta").Interface().(api.ObjectMeta)

			file, err = transformer.Print(objectMeta.Name, dirName, strings.ToLower(typeMeta.Kind), data, opt.ToStdout, opt.GenerateJSON, f)
			if err != nil {
				return errors.Wrap(err, "transformer.Print failed")
			}

			files = append(files, file)
		}
	}
	if opt.CreateChart {
		err = generateHelm(opt.InputFiles, files)
		if err != nil {
			return errors.Wrap(err, "generateHelm failed")
		}
	}
	return nil
}

// marshal object runtime.Object and return byte array
func marshal(obj runtime.Object, jsonFormat bool) (data []byte, err error) {
	// convert data to yaml or json
	if jsonFormat {
		data, err = json.MarshalIndent(obj, "", "  ")
	} else {
		data, err = yaml.Marshal(obj)
	}
	if err != nil {
		data = nil
	}
	return
}

// Convert object to versioned object
// if groupVersion is  empty (unversioned.GroupVersion{}), use version from original object (obj)
func convertToVersion(obj runtime.Object, groupVersion unversioned.GroupVersion) (runtime.Object, error) {

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
func (k *Kubernetes) PortsExist(name string, service kobject.ServiceConfig) bool {
	if len(service.Port) == 0 {
		log.Debugf("[%s] No ports defined. Headless service will be created.", name)
		return false
	}
	return true

}

// CreateService creates a k8s service
func (k *Kubernetes) CreateService(name string, service kobject.ServiceConfig, objects []runtime.Object) *api.Service {
	svc := k.InitSvc(name, service)

	// Configure the service ports.
	servicePorts := k.ConfigServicePorts(name, service)
	svc.Spec.Ports = servicePorts

	svc.Spec.Type = api.ServiceType(service.ServiceType)

	// Configure annotations
	annotations := transformer.ConfigAnnotations(service)
	svc.ObjectMeta.Annotations = annotations

	return svc
}

// CreateHeadlessService creates a k8s headless service.
// Thi is used for docker-compose services without ports. For such services we can't create regular Kubernetes Service.
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
func (k *Kubernetes) UpdateKubernetesObjects(name string, service kobject.ServiceConfig, objects *[]runtime.Object) error {
	// Configure the environment variables.
	envs := k.ConfigEnvs(name, service)

	// Configure the container volumes.
	volumesMount, volumes, pvc, err := k.ConfigVolumes(name, service)
	if err != nil {
		return errors.Wrap(err, "k.ConfigVolumes failed")
	}
	// Configure Tmpfs
	if len(service.TmpFs) > 0 {
		TmpVolumesMount, TmpVolumes := k.ConfigTmpfs(name, service)

		for _, volume := range TmpVolumes {
			volumes = append(volumes, volume)
		}
		for _, vMount := range TmpVolumesMount {
			volumesMount = append(volumesMount, vMount)
		}

	}

	if pvc != nil {
		// Looping on the slice pvc instead of `*objects = append(*objects, pvc...)`
		// because the type of objects and pvc is different, but when doing append
		// one element at a time it gets converted to runtime.Object for objects slice
		for _, p := range pvc {
			*objects = append(*objects, p)
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
			template.Spec.Containers[0].Name = service.ContainerName
		}
		template.Spec.Containers[0].Env = envs
		template.Spec.Containers[0].Command = service.Command
		template.Spec.Containers[0].Args = service.Args
		template.Spec.Containers[0].WorkingDir = service.WorkingDir
		template.Spec.Containers[0].VolumeMounts = volumesMount
		template.Spec.Containers[0].Stdin = service.Stdin
		template.Spec.Containers[0].TTY = service.Tty
		template.Spec.Volumes = volumes

		if service.StopGracePeriod != "" {
			template.Spec.TerminationGracePeriodSeconds, err = DurationStrToSecondsInt(service.StopGracePeriod)
			if err != nil {
				log.Warningf("Failed to parse duration \"%v\" for service \"%v\"", service.StopGracePeriod, name)
			}
		}

		// Configure the resource limits
		if service.MemLimit != 0 {
			memoryResourceList := api.ResourceList{
				api.ResourceMemory: *resource.NewQuantity(
					int64(service.MemLimit), "RandomStringForFormat")}
			template.Spec.Containers[0].Resources.Limits = memoryResourceList
		}

		podSecurityContext := &api.PodSecurityContext{}
		//set pid namespace mode
		if service.Pid != "" {
			if service.Pid == "host" {
				podSecurityContext.HostPID = true
			} else {
				log.Warningf("Ignoring PID key for service \"%v\". Invalid value \"%v\".", name, service.Pid)
			}
		}

		// Setup security context
		securityContext := &api.SecurityContext{}
		if service.Privileged == true {
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
		template.ObjectMeta.Labels = transformer.ConfigLabels(name)

		// Configure the container restart policy.
		switch service.Restart {
		case "", "always":
			template.Spec.RestartPolicy = api.RestartPolicyAlways
		case "no":
			template.Spec.RestartPolicy = api.RestartPolicyNever
		case "on-failure":
			template.Spec.RestartPolicy = api.RestartPolicyOnFailure
		default:
			return errors.New("Unknown restart policy " + service.Restart + " for service" + name)
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

func (k *Kubernetes) findDependentVolumes(svcname string, komposeObject kobject.KomposeObject) (volumes []api.Volume, volumeMounts []api.VolumeMount, err error) {
	// Get all the volumes and volumemounts this particular service is dependent on
	for _, dependentSvc := range komposeObject.ServiceConfigs[svcname].VolumesFrom {
		vols, volMounts, err := k.findDependentVolumes(dependentSvc, komposeObject)
		if err != nil {
			err = errors.Wrap(err, "k.findDependentVolumes failed")
			return nil, nil, err
		}
		volumes = append(volumes, vols...)
		volumeMounts = append(volumeMounts, volMounts...)
	}
	// add the volumes info of this service
	volMounts, vols, _, err := k.ConfigVolumes(svcname, komposeObject.ServiceConfigs[svcname])
	if err != nil {
		err = errors.Wrap(err, "k.ConfigVolumes failed")
		return nil, nil, err
	}
	volumes = append(volumes, vols...)
	volumeMounts = append(volumeMounts, volMounts...)
	return volumes, volumeMounts, nil
}

// VolumesFrom creates volums and volumeMounts for volumes_from
func (k *Kubernetes) VolumesFrom(objects *[]runtime.Object, komposeObject kobject.KomposeObject) error {
	for _, obj := range *objects {
		switch t := obj.(type) {
		case *api.ReplicationController:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts, err := k.findDependentVolumes(dependentSvc, komposeObject)
				if err != nil {
					return errors.Wrap(err, "k.findDependentVolumes")
				}
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		case *extensions.Deployment:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts, err := k.findDependentVolumes(dependentSvc, komposeObject)
				if err != nil {
					return errors.Wrap(err, "k.findDependentVolumes")
				}
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		case *extensions.DaemonSet:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts, err := k.findDependentVolumes(dependentSvc, komposeObject)
				if err != nil {
					return errors.Wrap(err, "k.findDependentVolumes")
				}
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		case *deployapi.DeploymentConfig:
			svcName := t.ObjectMeta.Name
			for _, dependentSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
				volumes, volumeMounts, err := k.findDependentVolumes(dependentSvc, komposeObject)
				if err != nil {
					return errors.Wrap(err, "k.findDependentVolumes")
				}
				t.Spec.Template.Spec.Volumes = append(t.Spec.Template.Spec.Volumes, volumes...)
				t.Spec.Template.Spec.Containers[0].VolumeMounts = append(t.Spec.Template.Spec.Containers[0].VolumeMounts, volumeMounts...)
			}
		}
	}
	return nil
}

//Ensure the kubernetes objects are in a consistent order
func SortedKeys(komposeObject kobject.KomposeObject) []string {
	var sortedKeys []string
	for name := range komposeObject.ServiceConfigs {
		sortedKeys = append(sortedKeys, name)
	}
	sort.Strings(sortedKeys)
	return sortedKeys
}

//converts duration string to *int64 in seconds
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
