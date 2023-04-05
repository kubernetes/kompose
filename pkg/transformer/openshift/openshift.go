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

package openshift

import (
	"fmt"
	"os"
	"sort"

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"
	deployapi "github.com/openshift/api/apps/v1"
	buildapi "github.com/openshift/api/build/v1"
	imageapi "github.com/openshift/api/image/v1"
	routeapi "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// OpenShift implements Transformer interface and represents OpenShift transformer
type OpenShift struct {
	// Anonymous field allows for inheritance. We are basically inheriting
	// all of kubernetes.Kubernetes Methods and variables here. We'll overwrite
	// some of those methods with our own for openshift.
	kubernetes.Kubernetes
}

// list of all unsupported keys for this transformer
// Keys are names of variables in kobject struct.
// this is map to make searching for keys easier
// to make sure that unsupported key is not going to be reported twice
// by keeping record if already saw this key in another service
var unsupportedKey = map[string]bool{}

// initImageStream initializes ImageStream object
func (o *OpenShift) initImageStream(name string, service kobject.ServiceConfig, opt kobject.ConvertOptions) *imageapi.ImageStream {
	if service.Image == "" {
		service.Image = name
	}
	// Retrieve tags and image name for mapping
	var importPolicy imageapi.TagImportPolicy
	if opt.InsecureRepository {
		importPolicy = imageapi.TagImportPolicy{Insecure: true}
	}

	var tags []imageapi.TagReference

	if service.Build != "" || opt.Build != "build-config" {
		tags = append(tags,
			imageapi.TagReference{
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: service.Image,
				},
				ImportPolicy: importPolicy,
				Name:         GetImageTag(service.Image),
			})
	}

	is := &imageapi.ImageStream{
		TypeMeta: kapi.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "image.openshift.io/v1",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: imageapi.ImageStreamSpec{
			Tags: tags,
		},
	}
	return is
}

func initBuildConfig(name string, service kobject.ServiceConfig, repo string, branch string) (*buildapi.BuildConfig, error) {
	contextDir, err := GetAbsBuildContext(service.Build)
	envList := transformer.EnvSort{}
	for envName, envValue := range service.BuildArgs {
		if *envValue == "\x00" {
			*envValue = os.Getenv(envName)
		}
		envList = append(envList, corev1.EnvVar{Name: envName, Value: *envValue})
	}
	// Stable sorts data while keeping the original order of equal elements
	// we need this because envs are not populated in any random order
	// this sorting ensures they are populated in a particular order
	sort.Stable(envList)
	if err != nil {
		return nil, errors.Wrap(err, name+"buildconfig cannot be created due to error in creating build context, getAbsBuildContext failed")
	}

	bc := &buildapi.BuildConfig{
		TypeMeta: kapi.TypeMeta{
			Kind:       "BuildConfig",
			APIVersion: "v1",
		},

		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: buildapi.BuildConfigSpec{
			Triggers: []buildapi.BuildTriggerPolicy{
				{Type: "ConfigChange"},
			},
			RunPolicy: "Serial",
			CommonSpec: buildapi.CommonSpec{
				Source: buildapi.BuildSource{
					Git: &buildapi.GitBuildSource{
						Ref: branch,
						URI: repo,
					},
					ContextDir: contextDir,
				},
				Strategy: buildapi.BuildStrategy{
					DockerStrategy: &buildapi.DockerBuildStrategy{
						DockerfilePath: service.Dockerfile,
						Env:            envList,
					},
				},
				Output: buildapi.BuildOutput{
					To: &corev1.ObjectReference{
						Kind: "ImageStreamTag",
						Name: name + ":" + GetImageTag(service.Image),
					},
				},
			},
		},
	}
	return bc, nil
}

// initDeploymentConfig initializes OpenShifts DeploymentConfig object
func (o *OpenShift) initDeploymentConfig(name string, service kobject.ServiceConfig, replicas int) *deployapi.DeploymentConfig {
	containerName := []string{name}

	// Properly add tags to the image name
	tag := GetImageTag(service.Image)

	// Use ContainerName if it was set
	if service.ContainerName != "" {
		containerName = []string{service.ContainerName}
	}

	var podSpec corev1.PodSpec
	if len(service.Configs) > 0 {
		podSpec = o.InitPodSpecWithConfigMap(name, " ", service)
	} else {
		podSpec = o.InitPodSpec(name, " ", "")
	}

	dc := &deployapi.DeploymentConfig{
		TypeMeta: kapi.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: int32(replicas),
			Selector: transformer.ConfigLabels(name),
			//UniqueLabelKey: p.Name,
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels: transformer.ConfigLabels(name),
				},
				Spec: podSpec,
			},
			Triggers: []deployapi.DeploymentTriggerPolicy{
				// Trigger new deploy when DeploymentConfig is created (config change)
				{
					Type: deployapi.DeploymentTriggerOnConfigChange,
				},
				{
					Type: deployapi.DeploymentTriggerOnImageChange,
					ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
						//Automatic - if new tag is detected - update image update inside the pod template
						Automatic:      true,
						ContainerNames: containerName,
						From: corev1.ObjectReference{
							Name: name + ":" + tag,
							Kind: "ImageStreamTag",
						},
					},
				},
			},
		},
	}

	update := service.GetOSUpdateStrategy()
	if update != nil {
		dc.Spec.Strategy = deployapi.DeploymentStrategy{
			Type:          deployapi.DeploymentStrategyTypeRolling,
			RollingParams: update,
		}
		log.Debugf("Set deployment '%s' rolling update: MaxSurge: %s, MaxUnavailable: %s", name, update.MaxSurge.String(), update.MaxUnavailable.String())
	}

	return dc
}

func (o *OpenShift) initRoute(name string, service kobject.ServiceConfig, port int32) *routeapi.Route {
	route := &routeapi.Route{
		TypeMeta: kapi.TypeMeta{
			Kind:       "Route",
			APIVersion: "v1",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: routeapi.RouteSpec{
			Port: &routeapi.RoutePort{
				TargetPort: intstr.IntOrString{
					IntVal: port,
				},
			},
			To: routeapi.RouteTargetReference{
				Kind: "Service",
				Name: name,
			},
		},
	}

	if service.ExposeService != "true" {
		route.Spec.Host = service.ExposeService
	}
	return route
}

// Transform maps komposeObject to openshift objects
// returns objects that are already sorted in the way that Services are first
func (o *OpenShift) Transform(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) ([]runtime.Object, error) {
	noSupKeys := o.Kubernetes.CheckUnsupportedKey(&komposeObject, unsupportedKey)
	for _, keyName := range noSupKeys {
		log.Warningf("OpenShift provider doesn't support %s key - ignoring", keyName)
	}
	// this will hold all the converted data
	var allobjects []runtime.Object
	var err error
	var composeFileDir string
	buildRepo := opt.BuildRepo
	buildBranch := opt.BuildBranch

	if komposeObject.Secrets != nil {
		secrets, err := o.CreateSecrets(komposeObject)
		if err != nil {
			return nil, errors.Wrapf(err, "create secrets error")
		}
		for _, item := range secrets {
			allobjects = append(allobjects, item)
		}
	}

	sortedKeys := kubernetes.SortedKeys(komposeObject)
	for _, name := range sortedKeys {
		service := komposeObject.ServiceConfigs[name]
		var objects []runtime.Object

		//replicas
		var replica int
		if opt.IsReplicaSetFlag || service.Replicas == 0 {
			replica = opt.Replicas
		} else {
			replica = service.Replicas
		}

		// If Deploy.Mode = Global has been set, make replica = 1 when generating DeploymentConfig
		if service.DeployMode == "global" {
			replica = 1
		}

		// Must build the images before conversion (got to add service.Image in case 'image' key isn't provided
		// Check to see if there is an InputFile (required!) before we build the container
		// Check that there's actually a Build key
		// Lastly, we must have an Image name to continue
		if opt.Build == "local" && opt.InputFiles != nil && service.Build != "" {
			// If there's no "image" key, use the name of the container that's built
			if service.Image == "" {
				service.Image = name
			}

			if service.Image == "" {
				return nil, fmt.Errorf("image key required within build parameters in order to build and push service '%s'", name)
			}

			// Build the container!
			err := transformer.BuildDockerImage(service, name)
			if err != nil {
				log.Fatalf("Unable to build Docker container for service %v: %v", name, err)
			}

			// Push the built container to the repo!
			err = transformer.PushDockerImageWithOpt(service, name, opt)
			if err != nil {
				log.Fatalf("Unable to push Docker image for service %v: %v", name, err)
			}
		}

		// Generate pod only and nothing more
		if service.Restart == "no" || service.Restart == "on-failure" {
			// Error out if Controller Object is specified with restart: 'on-failure'
			if opt.IsDeploymentConfigFlag {
				return nil, errors.New("Controller object cannot be specified with restart: 'on-failure'")
			}
			pod := o.InitPod(name, service)
			objects = append(objects, pod)
		} else {
			objects = o.CreateWorkloadAndConfigMapObjects(name, service, opt)

			if opt.CreateDeploymentConfig {
				objects = append(objects, o.initDeploymentConfig(name, service, replica)) // OpenShift DeploymentConfigs
				// create ImageStream after deployment (creating IS will trigger new deployment)
				objects = append(objects, o.initImageStream(name, service, opt))
			}

			// buildconfig needs to be added to objects after imagestream because of this Openshift bug: https://github.com/openshift/origin/issues/4518
			// Generate BuildConfig if the parameter has been passed
			if service.Build != "" && opt.Build == "build-config" {
				// Get the compose file directory
				composeFileDir, err = transformer.GetComposeFileDir(opt.InputFiles)
				if err != nil {
					log.Warningf("Error %v in detecting compose file's directory.", err)
					continue
				}

				// Check for Git
				if !HasGitBinary() && (buildRepo == "" || buildBranch == "") {
					return nil, errors.New("Git is not installed! Please install Git to create buildconfig, else supply source repository and branch to use for build using '--build-repo', '--build-branch' options respectively")
				}

				// Check the Git branch
				if buildBranch == "" {
					buildBranch, err = GetGitCurrentBranch(composeFileDir)
					if err != nil {
						return nil, errors.Wrap(err, "Buildconfig cannot be created because current git branch couldn't be detected.")
					}
				}

				// Detect the remote branches
				if opt.BuildRepo == "" {
					if err != nil {
						return nil, errors.Wrap(err, "Buildconfig cannot be created because remote for current git branch couldn't be detected.")
					}
					buildRepo, err = GetGitCurrentRemoteURL(composeFileDir)
					if err != nil {
						return nil, errors.Wrap(err, "Buildconfig cannot be created because git remote origin repo couldn't be detected.")
					}
				}

				// Initialize and build BuildConfig
				bc, err := initBuildConfig(name, service, buildRepo, buildBranch)
				if err != nil {
					return nil, errors.Wrap(err, "initBuildConfig failed")
				}
				objects = append(objects, bc) // Openshift BuildConfigs

				// Log what we're doing
				log.Infof("Buildconfig using %s::%s as source.", buildRepo, buildBranch)
			}
		}

		if o.PortsExist(service) {
			if service.ServiceType == "LoadBalancer" {
				svcs := o.CreateLBService(name, service)
				for _, svc := range svcs {
					svc.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyType(service.ServiceExternalTrafficPolicy)
					objects = append(objects, svc)
				}
				if len(svcs) > 1 {
					log.Warningf("Create multiple service to avoid using mixed protocol in the same service when it's loadbalancer type")
				}
			} else {
				svc := o.CreateService(name, service)
				objects = append(objects, svc)

				if service.ExposeService != "" {
					objects = append(objects, o.initRoute(name, service, svc.Spec.Ports[0].Port))
				}
				if service.ServiceExternalTrafficPolicy != "" && svc.Spec.Type != corev1.ServiceTypeNodePort {
					log.Warningf("External Traffic Policy is ignored for the service %v of type %v", name, service.ServiceType)
				}
			}
		} else if service.ServiceType == "Headless" {
			svc := o.CreateHeadlessService(name, service)
			objects = append(objects, svc)
			if service.ServiceExternalTrafficPolicy != "" {
				log.Warningf("External Traffic Policy is ignored for the service %v of type Headless", name)
			}
		}

		err := o.UpdateKubernetesObjects(name, service, opt, &objects)
		if err != nil {
			return nil, errors.Wrap(err, "Error transforming Kubernetes objects")
		}

		allobjects = append(allobjects, objects...)
	}

	// sort all object so Services are first
	o.SortServicesFirst(&allobjects)
	o.RemoveDupObjects(&allobjects)
	// o.FixWorkloadVersion(&allobjects)

	return allobjects, nil
}
