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

package app

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/docker/docker/api/client/bundlefile"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"

	"encoding/json"
	"io/ioutil"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	//client "k8s.io/kubernetes/pkg/client/unversioned"
	//cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"

	"github.com/fatih/structs"
	"github.com/ghodss/yaml"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

var unsupportedKey = map[string]string{
	"Build":         "",
	"CapAdd":        "",
	"CapDrop":       "",
	"CPUSet":        "",
	"CPUShares":     "",
	"ContainerName": "",
	"Devices":       "",
	"DNS":           "",
	"DNSSearch":     "",
	"Dockerfile":    "",
	"DomainName":    "",
	"Entrypoint":    "",
	"EnvFile":       "",
	"Hostname":      "",
	"LogDriver":     "",
	"MemLimit":      "",
	"MemSwapLimit":  "",
	"Net":           "",
	"Pid":           "",
	"Uts":           "",
	"Ipc":           "",
	"ReadOnly":      "",
	"StdinOpen":     "",
	"SecurityOpt":   "",
	"Tty":           "",
	"User":          "",
	"VolumeDriver":  "",
	"VolumesFrom":   "",
	"Expose":        "",
	"ExternalLinks": "",
	"LogOpt":        "",
	"ExtraHosts":    "",
}

// RandStringBytes generates randomly n-character string
func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// BeforeApp is an action that is executed before any cli command.
func BeforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

// Ps lists all rc, svc.
func Ps(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//if c.BoolT("svc") {
	//	fmt.Printf("%-20s%-20s%-20s%-20s\n", "Name", "Cluster IP", "Ports", "Selectors")
	//	for name := range p.Configs {
	//		var ports string
	//		var selectors string
	//		services, err := client.Services(api.NamespaceDefault).Get(name)
	//
	//		if err != nil {
	//			logrus.Debugf("Cannot find service for: ", name)
	//		} else {
	//
	//			for i := range services.Spec.Ports {
	//				p := strconv.Itoa(int(services.Spec.Ports[i].Port))
	//				ports += ports + string(services.Spec.Ports[i].Protocol) + "(" + p + "),"
	//			}
	//
	//			for k, v := range services.ObjectMeta.Labels {
	//				selectors += selectors + k + "=" + v + ","
	//			}
	//
	//			ports = strings.TrimSuffix(ports, ",")
	//			selectors = strings.TrimSuffix(selectors, ",")
	//
	//			fmt.Printf("%-20s%-20s%-20s%-20s\n", services.ObjectMeta.Name,
	//				services.Spec.ClusterIP, ports, selectors)
	//		}
	//
	//	}
	//}
	//
	//if c.BoolT("rc") {
	//	fmt.Printf("%-15s%-15s%-30s%-10s%-20s\n", "Name", "Containers", "Images",
	//		"Replicas", "Selectors")
	//	for name := range p.Configs {
	//		var selectors string
	//		var containers string
	//		var images string
	//		rc, err := client.ReplicationControllers(api.NamespaceDefault).Get(name)
	//
	//		/* Should grab controller, container, image, selector, replicas */
	//
	//		if err != nil {
	//			logrus.Debugf("Cannot find rc for: ", string(name))
	//		} else {
	//
	//			for k, v := range rc.Spec.Selector {
	//				selectors += selectors + k + "=" + v + ","
	//			}
	//
	//			for i := range rc.Spec.Template.Spec.Containers {
	//				c := rc.Spec.Template.Spec.Containers[i]
	//				containers += containers + c.Name + ","
	//				images += images + c.Image + ","
	//			}
	//			selectors = strings.TrimSuffix(selectors, ",")
	//			containers = strings.TrimSuffix(containers, ",")
	//			images = strings.TrimSuffix(images, ",")
	//
	//			fmt.Printf("%-15s%-15s%-30s%-10d%-20s\n", rc.ObjectMeta.Name, containers,
	//				images, rc.Spec.Replicas, selectors)
	//		}
	//	}
	//}

}

// Delete deletes all rc, svc.
func Delete(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//for name := range p.Configs {
	//	if len(c.String("name")) > 0 && name != c.String("name") {
	//		continue
	//	}
	//
	//	if c.BoolT("svc") {
	//		err := client.Services(api.NamespaceDefault).Delete(name)
	//		if err != nil {
	//			logrus.Fatalf("Unable to delete service %s: %s\n", name, err)
	//		}
	//	} else if c.BoolT("rc") {
	//		err := client.ReplicationControllers(api.NamespaceDefault).Delete(name)
	//		if err != nil {
	//			logrus.Fatalf("Unable to delete replication controller %s: %s\n", name, err)
	//		}
	//	}
	//}
}

// Scale scales rc.
func Scale(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//if c.Int("scale") <= 0 {
	//	logrus.Fatalf("Scale must be defined and a positive number")
	//}
	//
	//for name := range p.Configs {
	//	if len(c.String("rc")) == 0 || c.String("rc") == name {
	//		s, err := client.ExtensionsClient.Scales(api.NamespaceDefault).Get("ReplicationController", name)
	//		if err != nil {
	//			logrus.Fatalf("Error retrieving scaling data: %s\n", err)
	//		}
	//
	//		s.Spec.Replicas = int32(c.Int("scale"))
	//
	//		s, err = client.ExtensionsClient.Scales(api.NamespaceDefault).Update("ReplicationController", s)
	//		if err != nil {
	//			logrus.Fatalf("Error updating scaling data: %s\n", err)
	//		}
	//
	//		fmt.Printf("Scaling %s to: %d\n", name, s.Spec.Replicas)
	//	}
	//}
}

// Create the file to write to if --out is specified
func createOutFile(out string) *os.File {
	var f *os.File
	var err error
	if len(out) != 0 {
		f, err = os.Create(out)
		if err != nil {
			logrus.Fatalf("error opening file: %v", err)
		}
	}
	return f
}

// Init RC object
func initRC(name string, service ServiceConfig, replicas int) *api.ReplicationController {
	rc := &api.ReplicationController{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicationController",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
			//Labels: map[string]string{"service": name},
		},
		Spec: api.ReplicationControllerSpec{
			Replicas: int32(replicas),
			Selector: map[string]string{"service": name},
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
				//Labels: map[string]string{"service": name},
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

// Init SC object
func initSC(name string, service ServiceConfig) *api.Service {
	sc := &api.Service{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
			//Labels: map[string]string{"service": name},
		},
		Spec: api.ServiceSpec{
			Selector: map[string]string{"service": name},
		},
	}
	return sc
}

// Init DC object
func initDC(name string, service ServiceConfig) *extensions.Deployment {
	dc := &extensions.Deployment{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"service": name},
		},
		Spec: extensions.DeploymentSpec{
			Replicas: 1,
			Selector: &unversioned.LabelSelector{
				MatchLabels: map[string]string{"service": name},
			},
			//UniqueLabelKey: p.Name,
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: map[string]string{"service": name},
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
	return dc
}

// Init DS object
func initDS(name string, service ServiceConfig) *extensions.DaemonSet {
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
				ObjectMeta: api.ObjectMeta{
					Name: name,
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
	return ds
}

// Init RS object
func initRS(name string, service ServiceConfig) *extensions.ReplicaSet {
	rs := &extensions.ReplicaSet{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: extensions.ReplicaSetSpec{
			Replicas: 1,
			Selector: &unversioned.LabelSelector{
				MatchLabels: map[string]string{"service": name},
			},
			Template: api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{},
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
	return rs
}

// Configure the environment variables.
func configEnvs(name string, service ServiceConfig) []api.EnvVar {
	envs := []api.EnvVar{}
	for _, v := range service.Environment {
		envs = append(envs, api.EnvVar{
			Name:  v.Name,
			Value: v.Value,
		})
	}

	return envs
}

// Configure the container volumes.
func configVolumes(service ServiceConfig) ([]api.VolumeMount, []api.Volume) {
	volumesMount := []api.VolumeMount{}
	volumes := []api.Volume{}
	for _, volume := range service.Volumes {
		character := ":"
		if strings.Contains(volume, character) {
			hostDir := volume[0:strings.Index(volume, character)]
			hostDir = strings.TrimSpace(hostDir)
			containerDir := volume[strings.Index(volume, character)+1:]
			containerDir = strings.TrimSpace(containerDir)

			// check if ro/rw mode is defined
			readonly := true
			if strings.Index(volume, character) != strings.LastIndex(volume, character) {
				mode := volume[strings.LastIndex(volume, character)+1:]
				if strings.Compare(mode, "rw") == 0 {
					readonly = false
				}
				containerDir = containerDir[0:strings.Index(containerDir, character)]
			}

			// volumeName = random string of 20 chars
			volumeName := RandStringBytes(20)

			volumesMount = append(volumesMount, api.VolumeMount{Name: volumeName, ReadOnly: readonly, MountPath: containerDir})
			p := &api.HostPathVolumeSource{
				Path: hostDir,
			}
			//p.Path = hostDir
			volumeSource := api.VolumeSource{HostPath: p}
			volumes = append(volumes, api.Volume{Name: volumeName, VolumeSource: volumeSource})
		}
	}
	return volumesMount, volumes
}

// Configure the container ports.
func configPorts(name string, service ServiceConfig) []api.ContainerPort {
	ports := []api.ContainerPort{}
	for _, port := range service.Port {
		var p api.Protocol
		switch port.Protocol {
		default:
			p = api.ProtocolTCP
		case ProtocolTCP:
			p = api.ProtocolTCP
		case ProtocolUDP:
			p = api.ProtocolUDP
		}
		ports = append(ports, api.ContainerPort{
			HostPort:      port.HostPort,
			ContainerPort: port.ContainerPort,
			Protocol:      p,
		})
	}

	return ports
}

// Configure the container service ports.
func configServicePorts(name string, service ServiceConfig) []api.ServicePort {
	servicePorts := []api.ServicePort{}
	for _, port := range service.Port {
		var p api.Protocol
		switch port.Protocol {
		default:
			p = api.ProtocolTCP
		case ProtocolTCP:
			p = api.ProtocolTCP
		case ProtocolUDP:
			p = api.ProtocolUDP
		}
		var targetPort intstr.IntOrString
		targetPort.IntVal = port.HostPort
		targetPort.StrVal = strconv.Itoa(int(port.HostPort))
		servicePorts = append(servicePorts, api.ServicePort{
			Name:       strconv.Itoa(int(port.ContainerPort)),
			Protocol:   p,
			Port:       port.ContainerPort,
			TargetPort: targetPort,
		})
	}
	return servicePorts
}

// Transform data to json/yaml
func transformer(v interface{}, entity string, generateYaml bool) ([]byte, string) {
	// convert data to json / yaml
	data, err := json.MarshalIndent(v, "", "  ")
	if generateYaml == true {
		data, err = yaml.Marshal(v)
	}
	if err != nil {
		return nil, "Failed to marshal the " + entity
	}
	logrus.Debugf("%s\n", data)
	return data, ""
}

// load Environment Variable from bundles file
func loadEnvVars(service bundlefile.Service) ([]EnvVar, string) {
	envs := []EnvVar{}
	for _, env := range service.Env {
		character := "="
		if strings.Contains(env, character) {
			value := env[strings.Index(env, character)+1:]
			name := env[0:strings.Index(env, character)]
			name = strings.TrimSpace(name)
			value = strings.TrimSpace(value)
			envs = append(envs, EnvVar{
				Name:  name,
				Value: value,
			})
		} else {
			character = ":"
			if strings.Contains(env, character) {
				charQuote := "'"
				value := env[strings.Index(env, character)+1:]
				name := env[0:strings.Index(env, character)]
				name = strings.TrimSpace(name)
				value = strings.TrimSpace(value)
				if strings.Contains(value, charQuote) {
					value = strings.Trim(value, "'")
				}
				envs = append(envs, EnvVar{
					Name:  name,
					Value: value,
				})
			} else {
				return envs, "Invalid container env " + env
			}
		}
	}
	return envs, ""
}

// load Environment Variable from compose file
func loadEnvVarsFromCompose(e map[string]string) ([]EnvVar) {
	envs := []EnvVar{}
	for k, v := range e {
		envs = append(envs, EnvVar{
			Name:  k,
			Value: v,
		})
	}
	return envs
}

// load Ports from bundles file
func loadPorts(service bundlefile.Service) ([]Ports, string) {
	ports := []Ports{}
	for _, port := range service.Ports {
		var p Protocol
		switch port.Protocol {
		default:
			p = ProtocolTCP
		case "TCP":
			p = ProtocolTCP
		case "UDP":
			p = ProtocolUDP
		}
		ports = append(ports, Ports{
			HostPort:      int32(port.Port),
			ContainerPort: int32(port.Port),
			Protocol:      p,
		})
	}
	return ports, ""
}

// Load Ports from compose file
func loadPortsFromCompose(composePorts []string) ([]Ports, string) {
	ports := []Ports{}
	character := ":"
	for _, port := range composePorts {
		p := ProtocolTCP
		if strings.Contains(port, character) {
			hostPort := port[0:strings.Index(port, character)]
			hostPort = strings.TrimSpace(hostPort)
			hostPortInt, err := strconv.Atoi(hostPort)
			if err != nil {
				return nil, "Invalid host port of " + port
			}
			containerPort := port[strings.Index(port, character)+1:]
			containerPort = strings.TrimSpace(containerPort)
			containerPortInt, err := strconv.Atoi(containerPort)
			if err != nil {
				return nil, "Invalid container port of " + port
			}
			ports = append(ports, Ports{
				HostPort:      int32(hostPortInt),
				ContainerPort: int32(containerPortInt),
				Protocol:      p,
			})
		} else {
			containerPortInt, err := strconv.Atoi(port)
			if err != nil {
				return nil, "Invalid container port of " + port
			}
			ports = append(ports, Ports{
				ContainerPort: int32(containerPortInt),
				Protocol:      p,
			})
		}

	}
	return ports, ""
}

// load Image from bundles file
func loadImage(service bundlefile.Service) (string, string) {
	character := "@"
	if strings.Contains(service.Image, character) {
		return service.Image[0:strings.Index(service.Image, character)], ""
	}
	return "", "Invalid image format"
}

// Load DAB file into KomposeObject
func loadBundlesFile(file string) KomposeObject {
	komposeObject := KomposeObject{
		ServiceConfigs: make(map[string]ServiceConfig),
	}
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.Fatalf("Failed to read bundles file: ", err)
	}
	reader := strings.NewReader(string(buf))
	bundle, err := bundlefile.LoadFile(reader)
	if err != nil {
		logrus.Fatalf("Failed to parse bundles file: ", err)
	}

	for name, service := range bundle.Services {
		serviceConfig := ServiceConfig{}
		serviceConfig.Command = service.Command
		serviceConfig.Args = service.Args
		serviceConfig.Labels = service.Labels

		image, err := loadImage(service)
		if err != "" {
			logrus.Fatalf("Failed to load image from bundles file: " + err)
		}
		serviceConfig.Image = image

		envs, err := loadEnvVars(service)
		if err != "" {
			logrus.Fatalf("Failed to load envvar from bundles file: " + err)
		}
		serviceConfig.Environment = envs

		ports, err := loadPorts(service)
		if err != "" {
			logrus.Fatalf("Failed to load ports from bundles file: " + err)
		}
		serviceConfig.Port = ports

		if service.WorkingDir != nil {
			serviceConfig.WorkingDir = *service.WorkingDir
		}

		komposeObject.ServiceConfigs[name] = serviceConfig
	}
	return komposeObject
}

// Load compose file into KomposeObject
func loadComposeFile(file string, c *cli.Context) KomposeObject {
	komposeObject := KomposeObject{
		ServiceConfigs: make(map[string]ServiceConfig),
	}
	context := &docker.Context{}
	if file == "" {
		file = "docker-compose.yml"
	}
	context.ComposeFiles = []string{file}

	// load compose file into composeObject
	composeObject := project.NewProject(&context.Context, nil, nil)
	err := composeObject.Parse()
	if err != nil {
		logrus.Fatalf("Failed to load compose file", err)
	}

	// transform composeObject into komposeObject
	composeServiceNames := composeObject.ServiceConfigs.Keys()
	for _, name := range composeServiceNames {
		if composeServiceConfig, ok := composeObject.ServiceConfigs.Get(name); ok {
			// TODO: mapping composeObject config to komposeObject config
			serviceConfig := ServiceConfig{}
			serviceConfig.Image = composeServiceConfig.Image
			serviceConfig.ContainerName = composeServiceConfig.ContainerName

			// load environment variables
			envs := loadEnvVarsFromCompose(composeServiceConfig.Environment.ToMap())
			serviceConfig.Environment = envs

			// load ports
			ports, err := loadPortsFromCompose(composeServiceConfig.Ports)
			if err != "" {
				logrus.Fatalf("Failed to load ports from compose file: " + err)
			}
			serviceConfig.Port = ports

			serviceConfig.WorkingDir = composeServiceConfig.WorkingDir
			serviceConfig.Volumes = composeServiceConfig.Volumes

			// load labels
			labels := composeServiceConfig.Labels
			if labels != nil {
				if err := labels.UnmarshalYAML("", labels); err != nil {
					logrus.Fatalf("Failed to load labels from compose file: ", err)
				}
			}
			serviceConfig.Labels = labels

			serviceConfig.CPUSet = composeServiceConfig.CPUSet
			serviceConfig.CPUShares = composeServiceConfig.CPUShares
			serviceConfig.CPUQuota = composeServiceConfig.CPUQuota
			serviceConfig.CapAdd = composeServiceConfig.CapAdd
			serviceConfig.CapDrop = composeServiceConfig.CapDrop
			serviceConfig.Expose = composeServiceConfig.Expose
			serviceConfig.Privileged = composeServiceConfig.Privileged
			serviceConfig.Restart = composeServiceConfig.Restart
			serviceConfig.User = composeServiceConfig.User

			komposeObject.ServiceConfigs[name] = serviceConfig
		}
	}
	return komposeObject
}

// Convert komposeObject to K8S controllers
func komposeConvert(komposeObject KomposeObject, toStdout, createD, createRS, createDS, createChart, generateYaml bool, replicas int, inputFile string, outFile string, f *os.File) {
	mServices := make(map[string][]byte)
	mReplicationControllers := make(map[string][]byte)
	mDeployments := make(map[string][]byte)
	mDaemonSets := make(map[string][]byte)
	mReplicaSets := make(map[string][]byte)
	var svcnames []string

	for name, service := range komposeObject.ServiceConfigs {
		svcnames = append(svcnames, name)

		checkUnsupportedKey(service)

		rc := initRC(name, service, replicas)
		sc := initSC(name, service)
		dc := initDC(name, service)
		ds := initDS(name, service)
		rs := initRS(name, service)

		// Configure the environment variables.
		envs := configEnvs(name, service)

		// Configure the container command.
		var cmds []string
		for _, cmd := range service.Command {
			cmds = append(cmds, cmd)
		}
		// Configure the container volumes.
		volumesMount, volumes := configVolumes(service)

		// Configure the container ports.
		ports := configPorts(name, service)

		// Configure the service ports.
		servicePorts := configServicePorts(name, service)
		sc.Spec.Ports = servicePorts

		// Configure label
		labels := map[string]string{"service": name}
		for key, value := range service.Labels {
			labels[key] = value
		}
		sc.ObjectMeta.Labels = labels

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
		}

		// Update each supported controllers
		updateController(rc, fillTemplate, fillObjectMeta)
		updateController(rs, fillTemplate, fillObjectMeta)
		updateController(dc, fillTemplate, fillObjectMeta)
		updateController(ds, fillTemplate, fillObjectMeta)

		// convert datarc to json / yaml
		datarc, err := transformer(rc, "replication controller", generateYaml)
		if err != "" {
			logrus.Fatalf(err)
		}

		// convert datadc to json / yaml
		datadc, err := transformer(dc, "deployment", generateYaml)
		if err != "" {
			logrus.Fatalf(err)
		}

		// convert datads to json / yaml
		datads, err := transformer(ds, "daemonSet", generateYaml)
		if err != "" {
			logrus.Fatalf(err)
		}

		// convert datars to json / yaml
		datars, err := transformer(rs, "replicaSet", generateYaml)
		if err != "" {
			logrus.Fatalf(err)
		}

		// convert datasvc to json / yaml
		datasvc, err := transformer(sc, "service controller", generateYaml)
		if err != "" {
			logrus.Fatalf(err)
		}

		mServices[name] = datasvc
		mReplicationControllers[name] = datarc
		mDeployments[name] = datadc
		mDaemonSets[name] = datads
		mReplicaSets[name] = datars
	}

	for k, v := range mServices {
		if v != nil {
			print(k, "svc", v, toStdout, generateYaml, f)
		}
	}

	// If --out or --stdout is set, the validation should already prevent multiple controllers being generated
	if createD {
		for k, v := range mDeployments {
			print(k, "deployment", v, toStdout, generateYaml, f)
		}
	}

	if createDS {
		for k, v := range mDaemonSets {
			print(k, "daemonset", v, toStdout, generateYaml, f)
		}
	}

	if createRS {
		for k, v := range mReplicaSets {
			print(k, "replicaset", v, toStdout, generateYaml, f)
		}
	}

	if replicas != 0 {
		for k, v := range mReplicationControllers {
			print(k, "rc", v, toStdout, generateYaml, f)
		}
	}

	if f != nil {
		fmt.Fprintf(os.Stdout, "file %q created\n", outFile)
	}

	if createChart {
		err := generateHelm(inputFile, svcnames, generateYaml, createD, createDS, createRS, replicas)
		if err != nil {
			logrus.Fatalf("Failed to create Chart data: %s\n", err)
		}
	}
}

// Convert tranforms docker compose or dab file to k8s objects
func Convert(c *cli.Context) {
	inputFile := c.String("file")
	outFile := c.String("out")
	generateYaml := c.BoolT("yaml")
	toStdout := c.BoolT("stdout")
	createD := c.BoolT("deployment")
	createDS := c.BoolT("daemonset")
	createRS := c.BoolT("replicaset")
	createChart := c.BoolT("chart")
	fromBundles := c.BoolT("from-bundles")
	replicas := c.Int("replicationcontroller")
	singleOutput := len(outFile) != 0 || toStdout

	// Create Deployment by default if no controller has be set
	if !createD && !createDS && !createRS && replicas == 0 {
		createD = true
	}

	// Validate the flags
	if len(outFile) != 0 && toStdout {
		logrus.Fatalf("Error: --out and --stdout can't be set at the same time")
	}
	if createChart && toStdout {
		logrus.Fatalf("Error: chart cannot be generated when --stdout is specified")
	}
	if singleOutput {
		count := 0
		if createD {
			count++
		}
		if createDS {
			count++
		}
		if createRS {
			count++
		}
		if replicas != 0 {
			count++
		}
		if count > 1 {
			logrus.Fatalf("Error: only one type of Kubernetes controller can be generated when --out or --stdout is specified")
		}
	}

	var f *os.File
	if !createChart {
		f = createOutFile(outFile)
		defer f.Close()
	}

	komposeObject := KomposeObject{}

	if fromBundles {
		komposeObject = loadBundlesFile(inputFile)
	} else {
		komposeObject = loadComposeFile(inputFile, c)
	}

	// Convert komposeObject to K8S controllers
	komposeConvert(komposeObject, toStdout, createD, createRS, createDS, createChart, generateYaml, replicas, inputFile, outFile, f)
}

func checkUnsupportedKey(service ServiceConfig) {
	s := structs.New(service)
	for _, f := range s.Fields() {
		if f.IsExported() && !f.IsZero() {
			if _, ok := unsupportedKey[f.Name()]; ok {
				fmt.Println("WARNING: Unsupported key " + f.Name() + " - ignoring")
			}
		}
	}
}

func print(name, trailing string, data []byte, toStdout, generateYaml bool, f *os.File) {
	file := fmt.Sprintf("%s-%s.json", name, trailing)
	if generateYaml {
		file = fmt.Sprintf("%s-%s.yaml", name, trailing)
	}
	separator := ""
	if generateYaml {
		separator = "---"
	}
	if toStdout {
		fmt.Fprintf(os.Stdout, "%s%s\n", string(data), separator)
	} else if f != nil {
		// Write all content to a single file f
		if _, err := f.WriteString(fmt.Sprintf("%s%s\n", string(data), separator)); err != nil {
			logrus.Fatalf("Failed to write %s to file: %v", trailing, err)
		}
		f.Sync()
	} else {
		// Write content separately to each file
		if err := ioutil.WriteFile(file, []byte(data), 0644); err != nil {
			logrus.Fatalf("Failed to write %s: %v", trailing, err)
		}
		fmt.Fprintf(os.Stdout, "file %q created\n", file)
	}
}

// Up brings up rc, svc.
func Up(c *cli.Context) {
	//factory := cmdutil.NewFactory(nil)
	//clientConfig, err := factory.ClientConfig()
	//if err != nil {
	//	logrus.Fatalf("Failed to get Kubernetes client config: %v", err)
	//}
	//client := client.NewOrDie(clientConfig)
	//
	//files, err := ioutil.ReadDir(".")
	//if err != nil {
	//	logrus.Fatalf("Failed to load rc, svc manifest files: %s\n", err)
	//}
	//
	//// submit svc first
	//sc := &api.Service{}
	//for _, file := range files {
	//	if strings.Contains(file.Name(), "svc") {
	//		datasvc, err := ioutil.ReadFile(file.Name())
	//
	//		if err != nil {
	//			logrus.Fatalf("Failed to load %s: %s\n", file.Name(), err)
	//		}
	//
	//		if strings.Contains(file.Name(), "json") {
	//			err := json.Unmarshal(datasvc, &sc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to svc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		if strings.Contains(file.Name(), "yaml") {
	//			err := yaml.Unmarshal(datasvc, &sc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to svc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		// submit sc to k8s
	//		scCreated, err := client.Services(api.NamespaceDefault).Create(sc)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		logrus.Debugf("%s\n", scCreated)
	//	}
	//}
	//
	//// then submit rc
	//rc := &api.ReplicationController{}
	//for _, file := range files {
	//	if strings.Contains(file.Name(), "rc") {
	//		datarc, err := ioutil.ReadFile(file.Name())
	//
	//		if err != nil {
	//			logrus.Fatalf("Failed to load %s: %s\n", file.Name(), err)
	//		}
	//
	//		if strings.Contains(file.Name(), "json") {
	//			err := json.Unmarshal(datarc, &rc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to rc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		if strings.Contains(file.Name(), "yaml") {
	//			err := yaml.Unmarshal(datarc, &rc)
	//			if err != nil {
	//				logrus.Fatalf("Failed to unmarshal file %s to rc object: %s\n", file.Name(), err)
	//			}
	//		}
	//		// submit rc to k8s
	//		rcCreated, err := client.ReplicationControllers(api.NamespaceDefault).Create(rc)
	//		if err != nil {
	//			fmt.Println(err)
	//		}
	//		logrus.Debugf("%s\n", rcCreated)
	//	}
	//}

}

// updateController updates the given object with the given pod template update function and ObjectMeta update function
func updateController(obj runtime.Object, updateTemplate func(*api.PodTemplateSpec), updateMeta func(meta *api.ObjectMeta)) {
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
	}
}
