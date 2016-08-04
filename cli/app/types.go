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

// KomposeObject holds the generic struct of Kompose transformation
type KomposeObject struct {
	ServiceConfigs map[string]ServiceConfig
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
	Annotations   map[string]string
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
