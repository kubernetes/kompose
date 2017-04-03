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

package openshift

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kubernetes-incubator/kompose/pkg/kobject"
	"github.com/kubernetes-incubator/kompose/pkg/transformer/kubernetes"

	log "github.com/Sirupsen/logrus"

	"k8s.io/kubernetes/pkg/api"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/runtime"

	oclient "github.com/openshift/origin/pkg/client"
	ocliconfig "github.com/openshift/origin/pkg/cmd/cli/config"

	"time"

	"github.com/kubernetes-incubator/kompose/pkg/transformer"
	buildapi "github.com/openshift/origin/pkg/build/api"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deploymentconfigreaper "github.com/openshift/origin/pkg/deploy/cmd"
	imageapi "github.com/openshift/origin/pkg/image/api"
	routeapi "github.com/openshift/origin/pkg/route/api"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/intstr"
	"reflect"
)

// OpenShift implements Transformer interface and represents OpenShift transformer
type OpenShift struct {
	// Anonymous field allows for inheritance. We are basically inheriting
	// all of kubernetes.Kubernetes Methods and variables here. We'll overwite
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

// getImageTag get tag name from image name
// if no tag is specified return 'latest'
func getImageTag(image string) string {
	p := strings.Split(image, ":")
	if len(p) == 2 {
		return p[1]
	}
	return "latest"

}

// hasGitBinary checks if the 'git' binary is available on the system
func hasGitBinary() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// getGitCurrentRemoteURL gets current git remote URI for the current git repo
func getGitCurrentRemoteURL(composeFileDir string) (string, error) {
	cmd := exec.Command("git", "ls-remote", "--get-url")
	cmd.Dir = composeFileDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	url := strings.TrimRight(string(out), "\n")

	if !strings.HasSuffix(url, ".git") {
		url += ".git"
	}

	return url, nil
}

// getGitCurrentBranch gets current git branch name for the current git repo
func getGitCurrentBranch(composeFileDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = composeFileDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// getComposeFileDir returns compose file directory
func getComposeFileDir(inputFiles []string) (string, error) {
	// Lets assume all the docker-compose files are in the same directory
	inputFile := inputFiles[0]
	if strings.Index(inputFile, "/") != 0 {
		workDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		inputFile = filepath.Join(workDir, inputFile)
	}
	return filepath.Dir(inputFile), nil
}

// getAbsBuildContext returns build context relative to project root dir
func getAbsBuildContext(context string, composeFileDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
	cmd.Dir = composeFileDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	prefix := strings.Trim(string(out), "\n")
	return filepath.Join(prefix, context), nil
}

// initImageStream initialize ImageStream object
func (o *OpenShift) initImageStream(name string, service kobject.ServiceConfig) *imageapi.ImageStream {
	tag := getImageTag(service.Image)

	var tags map[string]imageapi.TagReference
	if service.Build == "" {
		tags = map[string]imageapi.TagReference{
			tag: imageapi.TagReference{
				From: &api.ObjectReference{
					Kind: "DockerImage",
					Name: service.Image,
				},
			},
		}
	}

	is := &imageapi.ImageStream{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "ImageStream",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: imageapi.ImageStreamSpec{
			Tags: tags,
		},
	}
	return is
}

// initBuildConfig initialize Openshifts BuildConfig Object
func initBuildConfig(name string, service kobject.ServiceConfig, composeFileDir string, repo string, branch string) (*buildapi.BuildConfig, error) {
	contextDir, err := getAbsBuildContext(service.Build, composeFileDir)
	if err != nil {
		return nil, errors.Wrap(err, name+"buildconfig cannot be created due to error in creating build context, getAbsBuildContext failed")
	}

	bc := &buildapi.BuildConfig{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "BuildConfig",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
		},
		Spec: buildapi.BuildConfigSpec{
			Triggers: []buildapi.BuildTriggerPolicy{
				{Type: "ConfigChange"},
				{Type: "ImageChange"},
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
					},
				},
				Output: buildapi.BuildOutput{
					To: &kapi.ObjectReference{
						Kind: "ImageStreamTag",
						Name: name + ":latest",
					},
				},
			},
		},
	}
	return bc, nil
}

// initDeploymentConfig initialize OpenShifts DeploymentConfig object
func (o *OpenShift) initDeploymentConfig(name string, service kobject.ServiceConfig, replicas int) *deployapi.DeploymentConfig {
	tag := getImageTag(service.Image)
	containerName := []string{name}

	// Use ContainerName if it was set
	if service.ContainerName != "" {
		containerName = []string{service.ContainerName}
	}

	dc := &deployapi.DeploymentConfig{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name:   name,
			Labels: transformer.ConfigLabels(name),
		},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: int32(replicas),
			Selector: transformer.ConfigLabels(name),
			//UniqueLabelKey: p.Name,
			Template: &api.PodTemplateSpec{
				ObjectMeta: api.ObjectMeta{
					Labels: transformer.ConfigLabels(name),
				},
				Spec: o.InitPodSpec(name, " "),
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
						From: api.ObjectReference{
							Name: name + ":" + tag,
							Kind: "ImageStreamTag",
						},
					},
				},
			},
		},
	}
	return dc
}

func (o *OpenShift) initRoute(name string, service kobject.ServiceConfig, port int32) *routeapi.Route {
	route := &routeapi.Route{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Route",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: name,
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
	hasBuild := false
	buildRepo := opt.BuildRepo
	buildBranch := opt.BuildBranch

	for name, service := range komposeObject.ServiceConfigs {
		var objects []runtime.Object

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
				objects = append(objects, o.initDeploymentConfig(name, service, opt.Replicas)) // OpenShift DeploymentConfigs
				// create ImageStream after deployment (creating IS will trigger new deployment)
				objects = append(objects, o.initImageStream(name, service))
			}

			// buildconfig needs to be added to objects after imagestream because of this Openshift bug: https://github.com/openshift/origin/issues/4518
			if service.Build != "" {
				if !hasBuild {
					composeFileDir, err = getComposeFileDir(opt.InputFiles)
					if err != nil {
						log.Warningf("Error in detecting compose file's directory.")
						continue
					}
					if !hasGitBinary() && (buildRepo == "" || buildBranch == "") {
						return nil, errors.New("Git is not installed! Please install Git to create buildconfig, else supply source repository and branch to use for build using '--build-repo', '--build-branch' options respectively")
					}
					if buildBranch == "" {
						buildBranch, err = getGitCurrentBranch(composeFileDir)
						if err != nil {
							return nil, errors.Wrap(err, "Buildconfig cannot be created because current git branch couldn't be detected.")
						}
					}
					if opt.BuildRepo == "" {
						if err != nil {
							return nil, errors.Wrap(err, "Buildconfig cannot be created because remote for current git branch couldn't be detected.")
						}
						buildRepo, err = getGitCurrentRemoteURL(composeFileDir)
						if err != nil {
							return nil, errors.Wrap(err, "Buildconfig cannot be created because git remote origin repo couldn't be detected.")
						}
					}
					hasBuild = true
				}
				bc, err := initBuildConfig(name, service, composeFileDir, buildRepo, buildBranch)
				if err != nil {
					return nil, errors.Wrap(err, "initBuildConfig failed")
				}
				objects = append(objects, bc) // Openshift BuildConfigs
			}

			// If ports not provided in configuration we will not make service
			if o.PortsExist(name, service) {
				svc := o.CreateService(name, service, objects)
				objects = append(objects, svc)

				if service.ExposeService != "" {
					objects = append(objects, o.initRoute(name, service, svc.Spec.Ports[0].Port))
				}
			} else {
				svc := o.CreateHeadlessService(name, service, objects)
				objects = append(objects, svc)
			}
		}
		o.UpdateKubernetesObjects(name, service, &objects)

		allobjects = append(allobjects, objects...)
	}

	if hasBuild {
		log.Infof("Buildconfig using %s::%s as source.", buildRepo, buildBranch)
	}
	// If docker-compose has a volumes_from directive it will be handled here
	o.VolumesFrom(&allobjects, komposeObject)
	// sort all object so Services are first
	o.SortServicesFirst(&allobjects)
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
	oclient := oclient.NewOrDie(oclientConfig)
	return oclient, nil
}

// Deploy transofrms and deploys kobject to OpenShift
func (o *OpenShift) Deploy(komposeObject kobject.KomposeObject, opt kobject.ConvertOptions) error {
	//Convert komposeObject
	objects, err := o.Transform(komposeObject, opt)

	if err != nil {
		return errors.Wrap(err, "o.Transform failed")
	}

	pvcStr := " "
	if !opt.EmptyVols {
		pvcStr = " and PersistentVolumeClaims "
	}
	fmt.Println("We are going to create OpenShift DeploymentConfigs, Services" + pvcStr + "for your Dockerized application. \n" +
		"If you need different kind of resources, use the 'kompose convert' and 'oc create -f' commands instead. \n")

	oclient, err := o.getOpenShiftClient()
	if err != nil {
		return err
	}
	kclient, namespace, err := o.GetKubernetesClient()
	if err != nil {
		return err
	}

	for _, v := range objects {
		switch t := v.(type) {
		case *imageapi.ImageStream:
			_, err := oclient.ImageStreams(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created ImageStream: %s", t.Name)
		case *buildapi.BuildConfig:
			_, err := oclient.BuildConfigs(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created BuildConfig: %s", t.Name)
		case *deployapi.DeploymentConfig:
			_, err := oclient.DeploymentConfigs(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created DeploymentConfig: %s", t.Name)
		case *api.Service:
			_, err := kclient.Services(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Service: %s", t.Name)
		case *api.PersistentVolumeClaim:
			_, err := kclient.PersistentVolumeClaims(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created PersistentVolumeClaim: %s", t.Name)
		case *routeapi.Route:
			_, err := oclient.Routes(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Route: %s", t.Name)
		case *api.Pod:
			_, err := kclient.Pods(namespace).Create(t)
			if err != nil {
				return err
			}
			log.Infof("Successfully created Pod: %s", t.Name)
		}
	}

	if !opt.EmptyVols {
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
	oclient, err := o.getOpenShiftClient()
	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}
	kclient, namespace, err := o.GetKubernetesClient()
	if err != nil {
		errorList = append(errorList, err)
		return errorList
	}

	for _, v := range objects {
		label := labels.SelectorFromSet(labels.Set(map[string]string{transformer.Selector: v.(meta.Object).GetName()}))
		options := api.ListOptions{LabelSelector: label}
		komposeLabel := map[string]string{transformer.Selector: v.(meta.Object).GetName()}
		switch t := v.(type) {
		case *imageapi.ImageStream:
			//delete imageStream
			imageStream, err := oclient.ImageStreams(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range imageStream.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = oclient.ImageStreams(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted ImageStream: %s", t.Name)
				}
			}

		case *buildapi.BuildConfig:
			//options := api.ListOptions{LabelSelector: label}
			buildConfig, err := oclient.BuildConfigs(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range buildConfig.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err := oclient.BuildConfigs(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted BuildConfig: %s", t.Name)
				}
			}

		case *deployapi.DeploymentConfig:
			// delete deploymentConfig
			deploymentConfig, err := oclient.DeploymentConfigs(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range deploymentConfig.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					dcreaper := deploymentconfigreaper.NewDeploymentConfigReaper(oclient, kclient)
					err := dcreaper.Stop(namespace, t.Name, TIMEOUT*time.Second, nil)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted DeploymentConfig: %s", t.Name)
				}
			}

		case *api.Service:
			//delete svc
			svc, err := kclient.Services(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range svc.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpService, err := kubectl.ReaperFor(api.Kind("Service"), kclient)
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

		case *api.PersistentVolumeClaim:
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
			route, err := oclient.Routes(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range route.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					err = oclient.Routes(namespace).Delete(t.Name)
					if err != nil {
						errorList = append(errorList, err)
						break
					}
					log.Infof("Successfully deleted Route: %s", t.Name)
				}
			}

		case *api.Pod:
			//delete pods
			pod, err := kclient.Pods(namespace).List(options)
			if err != nil {
				errorList = append(errorList, err)
				break
			}
			for _, l := range pod.Items {
				if reflect.DeepEqual(l.Labels, komposeLabel) {
					rpPod, err := kubectl.ReaperFor(api.Kind("Pod"), kclient)
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
