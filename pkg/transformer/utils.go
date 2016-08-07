package transformer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/skippbox/kompose/pkg/kobject"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/intstr"
	"github.com/skippbox/kompose/pkg/transformer/kubernetes"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"

// RandStringBytes generates randomly n-character string
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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
func InitRC(name string, service kobject.ServiceConfig, replicas int) *api.ReplicationController {
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
			Selector: map[string]string{"service": name},
			Replicas: int32(replicas),
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
func InitSC(name string, service kobject.ServiceConfig) *api.Service {
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
func InitDC(name string, service kobject.ServiceConfig, replicas int) *extensions.Deployment {
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
			Replicas: int32(replicas),
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

// Configure the container commands
func ConfigCommands(service kobject.ServiceConfig) []string {
	var cmds []string
	for _, cmd := range service.Command {
		cmds = append(cmds, cmd)
	}

	return cmds
}

// Configure the container volumes.
func ConfigVolumes(service kobject.ServiceConfig) ([]api.VolumeMount, []api.Volume) {
	volumesMount := []api.VolumeMount{}
	volumes := []api.Volume{}
	volumeSource := api.VolumeSource{}
	for _, volume := range service.Volumes {
		name, host, container, mode, err := parseVolume(volume)
		if err != nil {
			logrus.Warningf("Failed to configure container volume: %v", err)
			continue
		}

		// if volume name isn't specified, set it to a random string of 20 chars
		if len(name) == 0 {
			name = randStringBytes(20)
		}
		// check if ro/rw mode is defined, default rw
		readonly := len(mode) > 0 && mode == "ro"

		volumesMount = append(volumesMount, api.VolumeMount{Name: name, ReadOnly: readonly, MountPath: container})

		if len(host) > 0 {
			volumeSource = api.VolumeSource{HostPath: &api.HostPathVolumeSource{Path: host}}
		} else {
			volumeSource = api.VolumeSource{EmptyDir: &api.EmptyDirVolumeSource{}}
		}

		volumes = append(volumes, api.Volume{Name: name, VolumeSource: volumeSource})
	}
	return volumesMount, volumes
}

// parseVolume parse a given volume, which might be [name:][host:]container[:access_mode]
func parseVolume(volume string) (name, host, container, mode string, err error) {
	separator := ":"
	volumeStrings := strings.Split(volume, separator)
	if len(volumeStrings) == 0 {
		return
	}
	// Set name if existed
	if !isPath(volumeStrings[0]) {
		name = volumeStrings[0]
		volumeStrings = volumeStrings[1:]
	}
	if len(volumeStrings) == 0 {
		err = fmt.Errorf("invalid volume format: %s", volume)
		return
	}
	if volumeStrings[len(volumeStrings)-1] == "rw" || volumeStrings[len(volumeStrings)-1] == "ro" {
		mode = volumeStrings[len(volumeStrings)-1]
		volumeStrings = volumeStrings[:len(volumeStrings)-1]
	}
	container = volumeStrings[len(volumeStrings)-1]
	volumeStrings = volumeStrings[:len(volumeStrings)-1]
	if len(volumeStrings) == 1 {
		host = volumeStrings[0]
	}
	if !isPath(container) || (len(host) > 0 && !isPath(host)) || len(volumeStrings) > 1 {
		err = fmt.Errorf("invalid volume format: %s", volume)
		return
	}
	return
}

func isPath(substring string) bool {
	return strings.Contains(substring, "/")
}

// Configure the container ports.
func ConfigPorts(name string, service kobject.ServiceConfig) []api.ContainerPort {
	ports := []api.ContainerPort{}
	for _, port := range service.Port {
		var p api.Protocol
		switch port.Protocol {
		default:
			p = api.ProtocolTCP
		case kobject.ProtocolTCP:
			p = api.ProtocolTCP
		case kobject.ProtocolUDP:
			p = api.ProtocolUDP
		}
		ports = append(ports, api.ContainerPort{
			ContainerPort: port.ContainerPort,
			Protocol:      p,
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
		var p api.Protocol
		switch port.Protocol {
		default:
			p = api.ProtocolTCP
		case kobject.ProtocolTCP:
			p = api.ProtocolTCP
		case kobject.ProtocolUDP:
			p = api.ProtocolUDP
		}
		var targetPort intstr.IntOrString
		targetPort.IntVal = port.ContainerPort
		targetPort.StrVal = strconv.Itoa(int(port.ContainerPort))
		servicePorts = append(servicePorts, api.ServicePort{
			Name:       strconv.Itoa(int(port.HostPort)),
			Protocol:   p,
			Port:       port.HostPort,
			TargetPort: targetPort,
		})
	}
	return servicePorts
}

// Configure label
func ConfigLabels(name string) map[string]string {
	return map[string]string{"service": name}
}

// Configure annotations
func ConfigAnnotations(service kobject.ServiceConfig) map[string]string {
	annotations := map[string]string{}
	for key, value := range service.Annotations {
		annotations[key] = value
	}

	return annotations
}

// Transform data to json/yaml
func TransformData(obj runtime.Object, GenerateYaml bool) ([]byte, error) {
	//  Convert to versioned object
	objectVersion := obj.GetObjectKind().GroupVersionKind()
	version := unversioned.GroupVersion{Group: objectVersion.Group, Version: objectVersion.Version}
	versionedObj, err := api.Scheme.ConvertToVersion(obj, version)
	if err != nil {
		return nil, err
	}

	// convert data to json / yaml
	data, err := json.MarshalIndent(versionedObj, "", "  ")
	if GenerateYaml == true {
		data, err = yaml.Marshal(versionedObj)
	}
	if err != nil {
		return nil, err
	}
	logrus.Debugf("%s\n", data)
	return data, nil
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

func PrintControllers(mServices, mDeployments, mDaemonSets, mReplicationControllers, mDeploymentConfigs map[string][]byte, svcnames []string, opt kobject.ConvertOptions) {
	f := createOutFile(opt.OutFile)
	defer f.Close()

	for k, v := range mServices {
		if v != nil {
			print(k, "svc", v, opt.ToStdout, opt.GenerateYaml, f)
		}
	}

	// If --out or --stdout is set, the validation should already prevent multiple controllers being generated
	if opt.CreateD {
		for k, v := range mDeployments {
			print(k, "deployment", v, opt.ToStdout, opt.GenerateYaml, f)
		}
	}

	if opt.CreateDS {
		for k, v := range mDaemonSets {
			print(k, "daemonset", v, opt.ToStdout, opt.GenerateYaml, f)
		}
	}

	if opt.CreateRC {
		for k, v := range mReplicationControllers {
			print(k, "rc", v, opt.ToStdout, opt.GenerateYaml, f)
		}
	}

	if f != nil {
		fmt.Fprintf(os.Stdout, "file %q created\n", opt.OutFile)
	}

	if opt.CreateChart {
		err := kubernetes.GenerateHelm(opt.InputFile, svcnames, opt.GenerateYaml, opt.CreateD, opt.CreateDS, opt.CreateRC, opt.OutFile)
		if err != nil {
			logrus.Fatalf("Failed to create Chart data: %s\n", err)
		}
	}
}
