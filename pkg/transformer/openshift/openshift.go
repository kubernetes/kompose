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

	"github.com/kubernetes/kompose/pkg/kobject"
	"github.com/kubernetes/kompose/pkg/transformer/kubernetes"

	log "github.com/sirupsen/logrus"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/runtime"

	oclient "github.com/openshift/origin/pkg/client"
	ocliconfig "github.com/openshift/origin/pkg/cmd/cli/config"

	"time"

	"reflect"

	"sort"

	"github.com/kubernetes/kompose/pkg/transformer"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildconfigreaper "github.com/openshift/origin/pkg/build/cmd"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploymentconfigreaper "github.com/openshift/origin/pkg/deploy/cmd"
	imageapi "github.com/openshift/origin/pkg/image/api"
	routeapi "github.com/openshift/origin/pkg/route/api"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// OpenShift implements Transformer interface and represents OpenShift transformer
type OpenShift struct {
	// Anonymous field allows for inheritance. We are basically inheriting
	// all of kubernetes.Kubernetes Methods and variables here. We'll overwrite
	// some of those methods with our own for openshift.
	kubernetes.Kubernetes
}

// TIMEOUT is how long we'll wait for the termination of OpenShift resource to be successful
// used when undeploying resources from OpenShift
const TIMEOUT = 300

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
	tag := GetImageTag(service.Image)

	var importPolicy imageapi.TagImportPolicy
	if opt.InsecureRepository {
		importPolicy = imageapi.TagImportPolicy{Insecure: true}
	}

	var tags map[string]imageapi.TagReference

	if service.Build != "" || opt.Build != "build-config" {
		tags = map[string]imageapi.TagReference{
			tag: imageapi.TagReference{
				From: &kapi.ObjectReference{
					Kind: "DockerImage",
					Name: service.Image,
				},
				ImportPolicy: importPolicy,
			},
		}
	}

	is := &imageapi.ImageStream{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "v1",
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
		envList = append(envList, kapi.EnvVar{Name: envName, Value: *envValue})
	}
	// Stable sorts data while keeping the original order of equal elements
	// we need this because envs are not populated in any random order
	// this sorting ensures they are populated in a particular order
	sort.Stable(envList)
	if err != nil {
		return nil, errors.Wrap(err, name+"buildconfig cannot be created due to error in creating build context, getAbsBuildContext failed")
	}

	bc := &buildapi.BuildConfig{
		TypeMeta: unversioned.TypeMeta{
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
					To: &kapi.ObjectReference{
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

	var podSpec kapi.PodSpec
	if len(service.Configs) > 0 {
		podSpec = o.InitPodSpecWithConfigMap(name, " ", service)
	} else {
		podSpec = o.InitPodSpec(name, " ", "")
	}

	dc := &deployapi.DeploymentConfig{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "v1",
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: int32(replicas),
			Selector: transformer.ConfigLabels(name),
			//UniqueLabelKey: p.Name,
			Template: &kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{
					Labels: transformer.ConfigLabels(name),
				},
				Spec: podSpec,
			},
			Triggers: []deployapi.DeploymentTriggerPolicy{
				// Trigger new deploy when DeploymentConfig is created (config change)
				deployapi.DeploymentTriggerPolicy{
					Type: deployapi.DeploymentTriggerOnConfigChange,
				},
				deployapi.DeploymentTriggerPolicy{
					Type: deployapi.DeploymentTriggerOnImageChange,
					ImageChangeParams: &deployapi.DeploymentTriggerImageChangeParams{
						//Automatic - if new tag is detected - update image update inside the pod template
						Automatic:      true,
						ContainerNames: containerName,
						From: kapi.ObjectReference{
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
		TypeMeta: unversioned.TypeMeta{
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
			if opt.PushImage {
				err = transformer.PushDockerImage(service, name)
				if err != nil {
					log.Fatalf("Unable to push Docker image for service %v: %v", name, err)
				}
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
			objects = o.CreateKubernetesObjects(name, service, opt)

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
			svc := o.CreateService(name, service, objects)
			objects = append(objects, svc)

			if service.ExposeService != "" {
				objects = append(objects, o.initRoute(name, service, svc.Spec.Ports[0].Port))
			}
		} else if service.ServiceType == "Headless" {
			svc := o.CreateHeadlessService(name, service, objects)
			objects = append(objects, svc)
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
	o.FixWorkloadVersion(&allobjects)

	return allobjects, nil
}

// Create OpenShift client, returns OpenShift client
func (o *OpenShift) getOpenShiftClient() (*oclient.Client, error) {
	// initialize OpenShift Client
	loadingRules := ocliconfig.NewOpenShiftClientConfigLoadingRules()
	overrides := &clientcmd.ConfigOverrides{}
	oclientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, err
	}
	oc := oclient.NewOrDie(oclientConfig)
	return oc, nil
}

// Deploy transforms and deploys kobject to OpenShift
func (o *OpenShift) Deploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects, err := o.Transform(komposeObject, opt)

	if err != nil {
		return errors.Wrap(err, "o.Transform failed")
	}

	pvcStr := " "
	if !opt.EmptyVols || opt.Volumes != "emptyDir" {
		pvcStr = " and PersistentVolumeClaims "
	}
	log.Info("We are going to create OpenShift DeploymentConfigs, Services" + pvcStr + "for your Dockerized application. \n" +
		"If you need different kind of resources, use the 'kompose convert' and 'oc create -f' commands instead. \n")

	oc, err := o.getOpenShiftClient()
	if err != nil {
		return err
	}
	kclient, ns, err := o.GetKubernetesClient()
	if err != nil {
		return err
	}
	namespace := ns
	if opt.IsNamespaceFlag {
		namespace = opt.Namespace
	}

	log.Infof("Deploying application in %q namespace", namespace)

	for _, v := range objects {
		switch t := v.(type) {
		case *imageapi.ImageStream:
			_, err := oc.ImageStreams(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created ImageStream: %s", t.Name)
		case *buildapi.BuildConfig:
			_, err := oc.BuildConfigs(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created BuildConfig: %s", t.Name)
		case *deployapi.DeploymentConfig:
			_, err := oc.DeploymentConfigs(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created DeploymentConfig: %s", t.Name)
		case *kapi.Service:
			_, err := kclient.Services(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Service: %s", t.Name)
		case *kapi.PersistentVolumeClaim:
			_, err := kclient.PersistentVolumeClaims(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created PersistentVolumeClaim: %s of size %s. If your cluster has dynamic storage provisioning, you don't have to do anything. Otherwise you have to create PersistentVolume to make PVC work", t.Name, kubernetes.PVCRequestSize)
		case *routeapi.Route:
			_, err := oc.Routes(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Route: %s", t.Name)
		case *kapi.Pod:
			_, err := kclient.Pods(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Pod: %s", t.Name)
		}
	}

	if !opt.EmptyVols || opt.Volumes != "emptyDir" {
		pvcStr = ",pvc"
	} else {
		pvcStr = ""
	}
	fmt.Println("\nYour application has been deployed to OpenShift. You can run 'oc get dc,svc,is" + pvcStr + "' for details.")

	return nil
}

//Undeploy removes deployed artifacts from OpenShift cluster
func (o *OpenShift) Undeploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) []error {
	var errorList []error
	//Convert komposeObject
	objects, err := o.Transform(komposeObject, opt)

	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}
	oc, err := o.getOpenShiftClient()
	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}
	kclient, ns, err := o.GetKubernetesClient()
	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}
	namespace := ns
	if opt.IsNamespaceFlag {
		namespace = opt.Namespace
	}

	log.Infof("Deleting application in %q namespace", namespace)

	for _, v := range objects {
		label := labels.SelectorFromSet(labels.Set(map[string]string{transformer.Selector: v.(meta.Object).GetName()}))
		options := kapi.ListOptions{LabelSelector: label}
		komposeLabel := map[string]string{transformer.Selector: v.(meta.Object).GetName()}
		switch t := v.(type) {
		case *imageapi.ImageStream:
			//delete imageStream
			imageStream, err := oc.ImageStreams(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range imageStream.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = oc.ImageStreams(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted ImageStream: %s", t.Name)
				}
			}

		case *buildapi.BuildConfig:
			buildConfig, err := oc.BuildConfigs(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range buildConfig.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					bcreaper := buildconfigreaper.NewBuildConfigReaper(oc)
					err := bcreaper.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted BuildConfig: %s", t.Name)
				}
			}

		case *deployapi.DeploymentConfig:
			// delete deploymentConfig
			deploymentConfig, err := oc.DeploymentConfigs(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range deploymentConfig.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					dcreaper := deploymentconfigreaper.NewDeploymentConfigReaper(oc, kclient)
					err := dcreaper.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted DeploymentConfig: %s", t.Name)
				}
			}

		case *kapi.Service:
			//delete svc
			svc, err := kclient.Services(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range svc.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpService, err := kubectl.ReaperFor(kapi.Kind("Service"), kclient)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					//FIXME: gracePeriod is nil
					err = rpService.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Service: %s", t.Name)
				}
			}

		case *kapi.PersistentVolumeClaim:
			// delete pvc
			pvc, err := kclient.PersistentVolumeClaims(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range pvc.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = kclient.PersistentVolumeClaims(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted PersistentVolumeClaim: %s", t.Name)
				}
			}

		case *routeapi.Route:
			// delete route
			route, err := oc.Routes(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range route.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = oc.Routes(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Route: %s", t.Name)
				}
			}

		case *kapi.Pod:
			//delete pods
			pod, err := kclient.Pods(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range pod.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpPod, err := kubectl.ReaperFor(kapi.Kind("Pod"), kclient)
					if err != nil {
						errorList = append(errorList, err)
						break
					}

					//FIXME: gracePeriod is nil
					err = rpPod.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Pod: %s", t.Name)

				}
			}
		}
	}
	return errorList
}
