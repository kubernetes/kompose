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

package kobject

import (
	"github.com/Sirupsen/logrus"

	"fmt"
	"github.com/skippbox/kompose/pkg/loader"
	"github.com/fatih/structs"
	"github.com/skippbox/kompose/pkg/transformer"
)

var unsupportedKey = map[string]int{
	"Build":         0,
	"CapAdd":        0,
	"CapDrop":       0,
	"CPUSet":        0,
	"CPUShares":     0,
	"CPUQuota":      0,
	"CgroupParent":  0,
	"Devices":       0,
	"DependsOn":     0,
	"DNS":           0,
	"DNSSearch":     0,
	"DomainName":    0,
	"Entrypoint":    0,
	"EnvFile":       0,
	"Expose":        0,
	"Extends":       0,
	"ExternalLinks": 0,
	"ExtraHosts":    0,
	"Hostname":      0,
	"Ipc":           0,
	"Logging":       0,
	"MacAddress":    0,
	"MemLimit":      0,
	"MemSwapLimit":  0,
	"NetworkMode":   0,
	"Networks":      0,
	"Pid":           0,
	"SecurityOpt":   0,
	"ShmSize":       0,
	"StopSignal":    0,
	"VolumeDriver":  0,
	"VolumesFrom":   0,
	"Uts":           0,
	"ReadOnly":      0,
	"StdinOpen":     0,
	"Tty":           0,
	"User":          0,
	"Ulimits":       0,
	"Dockerfile":    0,
	"Net":           0,
	"Args":          0,
}

// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
	ServiceConfigs map[string]ServiceConfig
}

type ConvertOptions struct {
	ToStdout               bool
	CreateD                bool
	CreateRC               bool
	CreateDS               bool
	CreateDeploymentConfig bool
	CreateChart            bool
	GenerateYaml           bool
	Replicas               int
	InputFile              string
	OutFile                string
}

// ServiceConfig holds the basic struct of a container
type ServiceConfig struct {
	ContainerName string
	Image         string
	Environment   []EnvVar
	Port          []Ports
	Command       []string
	WorkingDir    string
	Args          []string
	Volumes       []string
	Network       []string
	Labels        map[string]string
	CPUSet        string
	CPUShares     int64
	CPUQuota      int64
	CapAdd        []string
	CapDrop       []string
	Entrypoint    []string
	Expose        []string
	Privileged    bool
	Restart       string
	User          string
}

// EnvVar holds the environment variable struct of a container
type EnvVar struct {
	Name  string
	Value string
}

// Ports holds the ports struct of a container
type Ports struct {
	HostPort      int32
	ContainerPort int32
	Protocol      Protocol
}

// Protocol defines network protocols supported for things like container ports.
type Protocol string

const (
	// ProtocolTCP is the TCP protocol.
	ProtocolTCP Protocol = "TCP"
	// ProtocolUDP is the UDP protocol.
	ProtocolUDP Protocol = "UDP"
)

// loader takes input and converts to KomposeObject
func (k *KomposeObject) Loader(file string, inp string) {
	switch inp {
	case "bundle":
		//k.loadBundleFile(file)
		loader.LoadBundle(k, file)
	case "compose":
		//k.loadComposeFile(file)
		loader.LoadCompose(k, file)
	default:
		logrus.Fatalf("Input file format is not supported")

	}
}

// transformer takes KomposeObject and converts to K8S / OpenShift primitives
func (k *KomposeObject) Transformer(opt ConvertOptions) {
	transformer.Transform(k, opt)
}

func CheckUnsupportedKey(service interface{}) {
	s := structs.New(service)
	for _, f := range s.Fields() {
		if f.IsExported() && !f.IsZero() {
			if count, ok := unsupportedKey[f.Name()]; ok && count == 0 {
				fmt.Println("WARNING: Unsupported key " + f.Name() + " - ignoring")
				unsupportedKey[f.Name()]++
			}
		}
	}
}
