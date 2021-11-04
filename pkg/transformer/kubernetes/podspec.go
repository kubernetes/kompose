package kubernetes

import (
	"reflect"
	"strconv"

	mapset "github.com/deckarep/golang-set"
	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PodSpec struct {
	api.PodSpec
}

type PodSpecOption func(*PodSpec)

func AddContainer(service kobject.ServiceConfig, opt kobject.ConvertOptions) PodSpecOption {
	return func(podSpec *PodSpec) {
		name := GetContainerName(service)
		image := service.Image

		if image == "" {
			image = name
		}

		envs, err := ConfigEnvs(service, opt)
		if err != nil {
			panic("Unable to load env variables")
		}

		podSpec.Containers = append(podSpec.Containers, api.Container{
			Name:           name,
			Image:          image,
			Env:            envs,
			Command:        service.Command,
			Args:           service.Args,
			WorkingDir:     service.WorkingDir,
			Stdin:          service.Stdin,
			TTY:            service.Tty,
			LivenessProbe:  configProbe(service.HealthChecks.Liveness),
			ReadinessProbe: configProbe(service.HealthChecks.Readiness),
		})

		podSpec.Affinity = ConfigAffinity(service)
	}
}

func TerminationGracePeriodSeconds(name string, service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		var err error
		if service.StopGracePeriod != "" {
			podSpec.TerminationGracePeriodSeconds, err = DurationStrToSecondsInt(service.StopGracePeriod)
			if err != nil {
				log.Warningf("Failed to parse duration \"%v\" for service \"%v\"", service.StopGracePeriod, name)
			}
		}
	}
}

// ResourcesLimits Configure the resource limits
func ResourcesLimits(service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		if service.MemLimit != 0 || service.CPULimit != 0 {
			resourceLimit := api.ResourceList{}

			if service.MemLimit != 0 {
				resourceLimit[api.ResourceMemory] = *resource.NewQuantity(int64(service.MemLimit), "RandomStringForFormat")
			}

			if service.CPULimit != 0 {
				resourceLimit[api.ResourceCPU] = *resource.NewMilliQuantity(service.CPULimit, resource.DecimalSI)
			}

			for i := range podSpec.Containers {
				podSpec.Containers[i].Resources.Limits = resourceLimit
			}
		}
	}
}

// ResourcesRequests Configure the resource requests
func ResourcesRequests(service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		if service.MemReservation != 0 || service.CPUReservation != 0 {
			resourceRequests := api.ResourceList{}

			if service.MemReservation != 0 {
				resourceRequests[api.ResourceMemory] = *resource.NewQuantity(int64(service.MemReservation), "RandomStringForFormat")
			}

			if service.CPUReservation != 0 {
				resourceRequests[api.ResourceCPU] = *resource.NewMilliQuantity(service.CPUReservation, resource.DecimalSI)
			}

			for i := range podSpec.Containers {
				podSpec.Containers[i].Resources.Requests = resourceRequests
			}
		}
	}
}

// SecurityContext Configure SecurityContext
func SecurityContext(name string, service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
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

		// Configure capabilities
		capabilities := ConfigCapabilities(service)

		//set capabilities if it is not empty
		if len(capabilities.Add) > 0 || len(capabilities.Drop) > 0 {
			securityContext.Capabilities = capabilities
		}

		// update template only if securityContext is not empty
		if *securityContext != (api.SecurityContext{}) {
			podSpec.Containers[0].SecurityContext = securityContext
		}
		if !reflect.DeepEqual(*podSecurityContext, api.PodSecurityContext{}) {
			podSpec.SecurityContext = podSecurityContext
		}
	}
}

func SetVolumeNames(volumes []api.Volume) mapset.Set {
	set := mapset.NewSet()
	for _, volume := range volumes {
		set.Add(volume.Name)
	}
	return set
}

func SetVolumes(volumes []api.Volume) PodSpecOption {
	return func(podSpec *PodSpec) {
		volumesSet := SetVolumeNames(volumes)
		containerVolumesSet := SetVolumeNames(podSpec.Volumes)
		for diffVolumeName := range volumesSet.Difference(containerVolumesSet).Iter() {
			for _, volume := range volumes {
				if volume.Name == diffVolumeName {
					podSpec.Volumes = append(podSpec.Volumes, volume)
					break
				}
			}
		}
	}
}

func SetVolumeMountPaths(volumesMount []api.VolumeMount) mapset.Set {
	set := mapset.NewSet()
	for _, volumeMount := range volumesMount {
		set.Add(volumeMount.MountPath)
	}

	return set
}

func SetVolumeMounts(volumesMount []api.VolumeMount) PodSpecOption {
	return func(podSpec *PodSpec) {
		volumesMountSet := SetVolumeMountPaths(volumesMount)
		for i := range podSpec.Containers {
			containerVolumeMountsSet := SetVolumeMountPaths(podSpec.Containers[i].VolumeMounts)
			for diffVolumeMountPath := range volumesMountSet.Difference(containerVolumeMountsSet).Iter() {
				for _, volumeMount := range volumesMount {
					if volumeMount.MountPath == diffVolumeMountPath {
						podSpec.Containers[i].VolumeMounts = append(podSpec.Containers[i].VolumeMounts, volumeMount)
						break
					}
				}
			}
		}
	}
}

// SetPorts Configure ports
func SetPorts(service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		// Configure the container ports.
		ports := ConfigPorts(service)

		for i := range podSpec.Containers {
			if GetContainerName(service) == podSpec.Containers[i].Name {
				podSpec.Containers[i].Ports = ports
			}
		}
	}
}

// ImagePullPolicy Configure the image pull policy
func ImagePullPolicy(name string, service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		if policy, err := GetImagePullPolicy(name, service.ImagePullPolicy); err != nil {
			panic(err)
		} else {
			for i := range podSpec.Containers {
				podSpec.Containers[i].ImagePullPolicy = policy
			}
		}
	}
}

// RestartPolicy Configure the container restart policy.
func RestartPolicy(name string, service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		if restart, err := GetRestartPolicy(name, service.Restart); err != nil {
			panic(err)
		} else {
			podSpec.RestartPolicy = restart
		}
	}
}

func HostName(service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		// Configure hostname/domain_name settings
		if service.HostName != "" {
			podSpec.Hostname = service.HostName
		}
	}
}

func DomainName(service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		if service.DomainName != "" {
			podSpec.Subdomain = service.DomainName
		}
	}
}

func configProbe(healthCheck kobject.HealthCheck) *api.Probe {
	probe := api.Probe{}
	// We check to see if it's blank or disable
	if reflect.DeepEqual(healthCheck, kobject.HealthCheck{}) || healthCheck.Disable {
		return nil
	}

	if len(healthCheck.Test) > 0 {
		probe.Handler = api.Handler{
			Exec: &api.ExecAction{
				Command: healthCheck.Test,
			},
		}
	} else if !reflect.ValueOf(healthCheck.HTTPPath).IsZero() && !reflect.ValueOf(healthCheck.HTTPPort).IsZero() {
		probe.Handler = api.Handler{
			HTTPGet: &api.HTTPGetAction{
				Path: healthCheck.HTTPPath,
				Port: intstr.FromInt(int(healthCheck.HTTPPort)),
			},
		}
	} else if !reflect.ValueOf(healthCheck.TCPPort).IsZero() {
		probe.Handler = api.Handler{
			TCPSocket: &api.TCPSocketAction{
				Port: intstr.FromInt(int(healthCheck.TCPPort)),
			},
		}
	} else {
		panic(errors.New("Health check must contain a command"))
	}

	probe.TimeoutSeconds = healthCheck.Timeout
	probe.PeriodSeconds = healthCheck.Interval
	probe.FailureThreshold = healthCheck.Retries

	// See issue: https://github.com/docker/cli/issues/116
	// StartPeriod has been added to v3.4 of the compose
	probe.InitialDelaySeconds = healthCheck.StartPeriod
	return &probe
}

func ServiceAccountName(serviceAccountName string) PodSpecOption {
	return func(podSpec *PodSpec) {
		podSpec.ServiceAccountName = serviceAccountName
	}
}

func TopologySpreadConstraints(service kobject.ServiceConfig) PodSpecOption {
	return func(podSpec *PodSpec) {
		podSpec.TopologySpreadConstraints = ConfigTopologySpreadConstraints(service)
	}
}

func (podSpec *PodSpec) Append(ops ...PodSpecOption) *PodSpec {
	for _, option := range ops {
		option(podSpec)
	}
	return podSpec
}

func (podSpec *PodSpec) Get() api.PodSpec {
	return podSpec.PodSpec
}
