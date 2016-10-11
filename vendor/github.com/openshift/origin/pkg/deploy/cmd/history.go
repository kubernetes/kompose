package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"text/tabwriter"

	kapi "k8s.io/kubernetes/pkg/api"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubectl"

	"github.com/openshift/origin/pkg/client"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
	deployutil "github.com/openshift/origin/pkg/deploy/util"
)

func NewDeploymentConfigHistoryViewer(oc client.Interface, kc kclient.Interface) kubectl.HistoryViewer {
	return &DeploymentConfigHistoryViewer{dn: oc, rn: kc}
}

// DeploymentConfigHistoryViewer is an implementation of the kubectl HistoryViewer interface
// for deployment configs.
type DeploymentConfigHistoryViewer struct {
	rn kclient.ReplicationControllersNamespacer
	dn client.DeploymentConfigsNamespacer
}

var _ kubectl.HistoryViewer = &DeploymentConfigHistoryViewer{}

// ViewHistory returns a description of all the history it can find for a deployment config.
func (h *DeploymentConfigHistoryViewer) ViewHistory(namespace, name string, revision int64) (string, error) {
	opts := kapi.ListOptions{LabelSelector: deployutil.ConfigSelector(name)}
	deploymentList, err := h.rn.ReplicationControllers(namespace).List(opts)
	if err != nil {
		return "", err
	}
	history := deploymentList.Items

	if len(deploymentList.Items) == 0 {
		return "No rollout history found.", nil
	}

	// Print details of a specific revision
	if revision > 0 {
		var desired *kapi.PodTemplateSpec
		// We could use a binary search here but brute-force is always faster to write
		for i := range history {
			rc := history[i]

			if deployutil.DeploymentVersionFor(&rc) == revision {
				desired = rc.Spec.Template
				break
			}
		}

		if desired == nil {
			return "", fmt.Errorf("unable to find the specified revision")
		}

		buf := bytes.NewBuffer([]byte{})
		kubectl.DescribePodTemplate(desired, buf)
		return buf.String(), nil
	}

	sort.Sort(deployutil.ByLatestVersionAsc(history))

	return tabbedString(func(out *tabwriter.Writer) error {
		fmt.Fprintf(out, "REVISION\tSTATUS\tCAUSE\n")
		for i := range history {
			rc := history[i]

			rev := deployutil.DeploymentVersionFor(&rc)
			status := deployutil.DeploymentStatusFor(&rc)
			cause := rc.Annotations[deployapi.DeploymentStatusReasonAnnotation]
			if len(cause) == 0 {
				cause = "<unknown>"
			}
			fmt.Fprintf(out, "%d\t%s\t%s\n", rev, status, cause)
		}
		return nil
	})
}

// TODO: Re-use from an utility package
func tabbedString(f func(*tabwriter.Writer) error) (string, error) {
	out := new(tabwriter.Writer)
	buf := &bytes.Buffer{}
	out.Init(buf, 0, 8, 1, '\t', 0)

	err := f(out)
	if err != nil {
		return "", err
	}

	out.Flush()
	str := string(buf.String())
	return str, nil
}
