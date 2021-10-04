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

package compose

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/docker/cli/opts"
	"github.com/docker/go-connections/nat"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	api "k8s.io/api/core/v1"
)

// Load ports from compose file
// also load `expose` here
func loadPorts(composePorts []string, expose []string) ([]kobject.Ports, error) {
	kp := []kobject.Ports{}
	exist := map[string]bool{}

	for _, cp := range composePorts {
		var hostIP string

		if parts := strings.Split(cp, ":"); len(parts) == 3 {
			if ip := net.ParseIP(parts[0]); ip.To4() == nil && ip.To16() == nil {
				return nil, fmt.Errorf("%q contains an invalid IPv4 or IPv6 IP address", parts[0])
			}
			hostIP = parts[0]
		}

		np, pbs, err := nat.ParsePortSpecs([]string{cp})
		if err != nil {
			return nil, fmt.Errorf("invalid port, error = %v", err)
		}
		// Force HostIP value to avoid warning raised by github.com/docker/cli/opts
		// The opts package will warn if the bindings contains host IP except
		// 0.0.0.0. However, the message is not useful in this case since the value
		// should be handled by kompose properly.
		for _, pb := range pbs {
			for i, p := range pb {
				p.HostIP = ""
				pb[i] = p
			}
		}

		var ports []string
		for p := range np {
			ports = append(ports, string(p))
		}
		sort.Strings(ports)

		for _, p := range ports {
			pc, err := opts.ConvertPortToPortConfig(nat.Port(p), pbs)
			if err != nil {
				return nil, fmt.Errorf("invalid port, error = %v", err)
			}
			for _, cfg := range pc {
				kp = append(kp, kobject.Ports{
					HostPort:      int32(cfg.PublishedPort),
					ContainerPort: int32(cfg.TargetPort),
					HostIP:        hostIP,
					Protocol:      strings.ToUpper(string(cfg.Protocol)),
				})
			}
		}
	}

	// load remain expose ports
	for _, p := range kp {
		// must use cast...
		exist[cast.ToString(p.ContainerPort)+p.Protocol] = true
	}

	if expose != nil {
		for _, port := range expose {
			portValue := port
			protocol := string(api.ProtocolTCP)
			if strings.Contains(portValue, "/") {
				splits := strings.Split(port, "/")
				portValue = splits[0]
				protocol = splits[1]
			}

			if !exist[portValue+protocol] {
				kp = append(kp, kobject.Ports{
					ContainerPort: cast.ToInt32(portValue),
					Protocol:      strings.ToUpper(protocol),
				})
			}
		}
	}

	return kp, nil
}

func checkLabelsPorts(noOfPort int, labels string, svcName string) error {
	if noOfPort == 0 && (labels == "NodePort" || labels == "LoadBalancer") {
		return errors.Errorf("%s defined in service %s with no ports present. Issues may occur when bringing up artifacts.", labels, svcName)
	}
	return nil
}

// returns all volumes associated with service, if `volumes_from` key is used, we have to retrieve volumes from the services which are mentioned there. Hence, recursive function is used here.
func retrieveVolume(svcName string, komposeObject kobject.KomposeObject) (volume []kobject.Volumes, err error) {
	// if volumes-from key is present
	if komposeObject.ServiceConfigs[svcName].VolumesFrom != nil {
		// iterating over services from `volumes-from`
		for _, depSvc := range komposeObject.ServiceConfigs[svcName].VolumesFrom {
			// recursive call for retrieving volumes of services from `volumes-from`
			dVols, err := retrieveVolume(depSvc, komposeObject)
			if err != nil {
				return nil, errors.Wrapf(err, "could not retrieve the volume")
			}
			var cVols []kobject.Volumes
			cVols, err = ParseVols(komposeObject.ServiceConfigs[svcName].VolList, svcName)
			if err != nil {
				return nil, errors.Wrapf(err, "error generating current volumes")
			}

			for _, cv := range cVols {
				// check whether volumes of current service is same or not as that of dependent volumes coming from `volumes-from`
				ok, dv := getVol(cv, dVols)
				if ok {
					// change current volumes service name to dependent service name
					if dv.VFrom == "" {
						cv.VFrom = dv.SvcName
						cv.SvcName = dv.SvcName
					} else {
						cv.VFrom = dv.VFrom
						cv.SvcName = dv.SvcName
					}
					cv.PVCName = dv.PVCName
				}
				volume = append(volume, cv)
			}
			// iterating over dependent volumes
			for _, dv := range dVols {
				// check whether dependent volume is already present or not
				if checkVolDependent(dv, volume) {
					// if found, add service name to `VFrom`
					dv.VFrom = dv.SvcName
					volume = append(volume, dv)
				}
			}
		}
	} else {
		// if `volumes-from` is not present
		volume, err = ParseVols(komposeObject.ServiceConfigs[svcName].VolList, svcName)
		if err != nil {
			return nil, errors.Wrapf(err, "error generating current volumes")
		}
	}
	return
}

// checkVolDependent returns false if dependent volume is present
func checkVolDependent(dv kobject.Volumes, volume []kobject.Volumes) bool {
	for _, vol := range volume {
		if vol.PVCName == dv.PVCName {
			return false
		}
	}
	return true
}

// ParseVols parse volumes
func ParseVols(volNames []string, svcName string) ([]kobject.Volumes, error) {
	var volumes []kobject.Volumes
	var err error

	for i, vn := range volNames {
		var v kobject.Volumes
		v.VolumeName, v.Host, v.Container, v.Mode, err = transformer.ParseVolume(vn)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse volume %q: %v", vn, err)
		}
		v.VolumeName = normalizeVolumes(v.VolumeName)
		v.SvcName = svcName
		v.MountPath = fmt.Sprintf("%s:%s", v.Host, v.Container)
		v.PVCName = fmt.Sprintf("%s-claim%d", v.SvcName, i)
		volumes = append(volumes, v)
	}

	return volumes, nil
}

// for dependent volumes, returns true and the respective volume if mountpath are same
func getVol(toFind kobject.Volumes, Vols []kobject.Volumes) (bool, kobject.Volumes) {
	for _, dv := range Vols {
		if toFind.MountPath == dv.MountPath {
			return true, dv
		}
	}
	return false, kobject.Volumes{}
}
