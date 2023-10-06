package client

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	appsv1 "k8s.io/api/apps/v1"
)

func TestConvertError(t *testing.T) {
	randomBuildValue := "random-build"
	randomVolumeTypeValue := "random-volume-type"
	randomKubernetesControllerValue := "random-controller"
	randomKubernetesServiceGroupModeValue := "random-group-mode"
	buildConfigValue := string(BUILD_CONFIG)
	testCases := []struct {
		options      ConvertOptions
		errorMessage string
	}{
		{
			options: ConvertOptions{
				Build: &randomBuildValue,
			},
			errorMessage: fmt.Sprintf("unexpected Value for Build field. Possible values are: %v, %v, and %v", string(LOCAL), string(BUILD_CONFIG), string(NONE)),
		},
		{
			options: ConvertOptions{
				VolumeType: &randomVolumeTypeValue,
			},
			errorMessage: fmt.Sprintf("unexpected Value for VolumeType field. Possible values are: %v, %v, %v, %v", string(PVC), string(EMPTYDIR), string(HOSTPATH), string(CONFIGMAP)),
		},
		{
			options: ConvertOptions{
				Provider: Kubernetes{
					Controller: &randomKubernetesControllerValue,
				},
			},
			errorMessage: fmt.Sprintf("unexpected Value for Kubernetes Controller field. Possible values are: %v, %v, and %v", string(DEPLOYMENT), string(DAEMONSET), string(REPLICATION_CONTROLLER)),
		},
		{
			options: ConvertOptions{
				Provider: Kubernetes{
					ServiceGroupMode: &randomKubernetesServiceGroupModeValue,
				},
			},
			errorMessage: fmt.Sprintf("unexpected Value for Kubernetes Service Groupe Mode field. Possible values are: %v, %v, ''", string(LABEL), string(VOLUME)),
		},
		{
			options: ConvertOptions{
				Provider: Kubernetes{},
				Build:    &buildConfigValue,
			},
			errorMessage: fmt.Sprintf("the build value %v is only supported for Openshift provider", string(BUILD_CONFIG)),
		},
	}

	client, err := NewClient()
	assert.Check(t, is.Equal(err, nil))
	for _, tc := range testCases {
		_, err := client.Convert(tc.options)

		assert.Check(t, is.Equal(err.Error(), tc.errorMessage))
	}
}

func TestConvertWithDefaultOptions(t *testing.T) {
	client, err := NewClient(WithErrorOnWarning())
	assert.Check(t, is.Equal(err, nil))
	objects, err := client.Convert(ConvertOptions{
		OutFile: "./testdata/generated/",
		InputFiles: []string{
			"./testdata/docker-compose.yaml",
		},
	})
	assert.Check(t, is.Equal(err, nil))
	for _, object := range objects {
		if deployment, ok := object.(*appsv1.Deployment); ok {
			assert.Check(t, is.Equal(int(*deployment.Spec.Replicas), 1))
		}
	}
}
