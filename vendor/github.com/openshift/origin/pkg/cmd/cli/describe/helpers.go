package describe

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	units "github.com/docker/go-units"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util/sets"

	buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/client"
	imageapi "github.com/openshift/origin/pkg/image/api"
)

const emptyString = "<none>"

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

func toString(v interface{}) string {
	value := fmt.Sprintf("%v", v)
	if len(value) == 0 {
		value = emptyString
	}
	return value
}

func bold(v interface{}) string {
	return "\033[1m" + toString(v) + "\033[0m"
}

func convertEnv(env []api.EnvVar) map[string]string {
	result := make(map[string]string, len(env))
	for _, e := range env {
		result[e.Name] = toString(e.Value)
	}
	return result
}

func formatEnv(env api.EnvVar) string {
	if env.ValueFrom != nil && env.ValueFrom.FieldRef != nil {
		return fmt.Sprintf("%s=<%s>", env.Name, env.ValueFrom.FieldRef.FieldPath)
	}
	return fmt.Sprintf("%s=%s", env.Name, env.Value)
}

func formatString(out *tabwriter.Writer, label string, v interface{}) {
	fmt.Fprintf(out, fmt.Sprintf("%s:\t%s\n", label, toString(v)))
}

func formatTime(out *tabwriter.Writer, label string, t time.Time) {
	fmt.Fprintf(out, fmt.Sprintf("%s:\t%s ago\n", label, formatRelativeTime(t)))
}

func formatLabels(labelMap map[string]string) string {
	return labels.Set(labelMap).String()
}

func extractAnnotations(annotations map[string]string, keys ...string) ([]string, map[string]string) {
	extracted := make([]string, len(keys))
	remaining := make(map[string]string)
	for k, v := range annotations {
		remaining[k] = v
	}
	for i, key := range keys {
		extracted[i] = remaining[key]
		delete(remaining, key)
	}
	return extracted, remaining
}

func formatMapStringString(out *tabwriter.Writer, label string, items map[string]string) {
	keys := sets.NewString()
	for k := range items {
		keys.Insert(k)
	}
	if keys.Len() == 0 {
		formatString(out, label, "")
		return
	}
	for i, key := range keys.List() {
		if i == 0 {
			formatString(out, label, fmt.Sprintf("%s=%s", key, items[key]))
		} else {
			fmt.Fprintf(out, "%s\t%s=%s\n", "", key, items[key])
		}
	}
}

func formatAnnotations(out *tabwriter.Writer, m api.ObjectMeta, prefix string) {
	values, annotations := extractAnnotations(m.Annotations, "description")
	if len(values[0]) > 0 {
		formatString(out, prefix+"Description", values[0])
	}
	formatMapStringString(out, prefix+"Annotations", annotations)
}

var timeNowFn = func() time.Time {
	return time.Now()
}

// Receives a time.Duration and returns Docker go-utils'
// human-readable output
func formatToHumanDuration(dur time.Duration) string {
	return units.HumanDuration(dur)
}

func formatRelativeTime(t time.Time) string {
	return units.HumanDuration(timeNowFn().Sub(t))
}

// FormatRelativeTime converts a time field into a human readable age string (hours, minutes, days).
func FormatRelativeTime(t time.Time) string {
	return formatRelativeTime(t)
}

func formatMeta(out *tabwriter.Writer, m api.ObjectMeta) {
	formatString(out, "Name", m.Name)
	formatString(out, "Namespace", m.Namespace)
	if !m.CreationTimestamp.IsZero() {
		formatTime(out, "Created", m.CreationTimestamp.Time)
	}
	formatMapStringString(out, "Labels", m.Labels)
	formatAnnotations(out, m, "")
}

// DescribeWebhook holds the URL information about a webhook and for generic
// webhooks it tells us if we allow env variables.
type DescribeWebhook struct {
	URL      string
	AllowEnv *bool
}

// webhookDescribe returns a map of webhook trigger types and its corresponding
// information.
func webHooksDescribe(triggers []buildapi.BuildTriggerPolicy, name, namespace string, cli client.BuildConfigsNamespacer) map[string][]DescribeWebhook {
	result := map[string][]DescribeWebhook{}

	for _, trigger := range triggers {
		var webHookTrigger string
		var allowEnv *bool

		switch trigger.Type {
		case buildapi.GitHubWebHookBuildTriggerType:
			webHookTrigger = trigger.GitHubWebHook.Secret

		case buildapi.GenericWebHookBuildTriggerType:
			webHookTrigger = trigger.GenericWebHook.Secret
			allowEnv = &trigger.GenericWebHook.AllowEnv

		default:
			continue
		}
		webHookDesc := result[string(trigger.Type)]

		if len(webHookTrigger) == 0 {
			continue
		}

		var urlStr string
		url, err := cli.BuildConfigs(namespace).WebHookURL(name, &trigger)
		if err != nil {
			urlStr = fmt.Sprintf("<error: %s>", err.Error())
		} else {
			urlStr = url.String()
		}

		webHookDesc = append(webHookDesc,
			DescribeWebhook{
				URL:      urlStr,
				AllowEnv: allowEnv,
			})
		result[string(trigger.Type)] = webHookDesc
	}

	return result
}

var reLongImageID = regexp.MustCompile(`[a-f0-9]{60,}$`)

// shortenImagePullSpec returns a version of the pull spec intended for
// display, which may result in the image not being usable via cut-and-paste
// for users.
func shortenImagePullSpec(spec string) string {
	if reLongImageID.MatchString(spec) {
		return spec[:len(spec)-50]
	}
	return spec
}

func formatImageStreamTags(out *tabwriter.Writer, stream *imageapi.ImageStream) {
	if len(stream.Status.Tags) == 0 && len(stream.Spec.Tags) == 0 {
		fmt.Fprintf(out, "Tags:\t<none>\n")
		return
	}

	now := timeNowFn()

	images := make(map[string]string)
	for tag, tags := range stream.Status.Tags {
		for _, item := range tags.Items {
			switch {
			case len(item.Image) > 0:
				if _, ok := images[item.Image]; !ok {
					images[item.Image] = tag
				}
			case len(item.DockerImageReference) > 0:
				if _, ok := images[item.DockerImageReference]; !ok {
					images[item.Image] = item.DockerImageReference
				}
			}
		}
	}

	sortedTags := []string{}
	for k := range stream.Status.Tags {
		sortedTags = append(sortedTags, k)
	}
	var localReferences sets.String
	var referentialTags map[string]sets.String
	for k := range stream.Spec.Tags {
		if target, _, ok, multiple := imageapi.FollowTagReference(stream, k); ok && multiple {
			if referentialTags == nil {
				referentialTags = make(map[string]sets.String)
			}
			if localReferences == nil {
				localReferences = sets.NewString()
			}
			localReferences.Insert(k)
			v := referentialTags[target]
			if v == nil {
				v = sets.NewString()
				referentialTags[target] = v
			}
			v.Insert(k)
		}
		if _, ok := stream.Status.Tags[k]; !ok {
			sortedTags = append(sortedTags, k)
		}
	}
	fmt.Fprintf(out, "Unique Images:\t%d\nTags:\t%d\n\n", len(images), len(sortedTags))

	first := true
	imageapi.PrioritizeTags(sortedTags)
	for _, tag := range sortedTags {
		if localReferences.Has(tag) {
			continue
		}
		if first {
			first = false
		} else {
			fmt.Fprintf(out, "\n")
		}
		taglist, _ := stream.Status.Tags[tag]
		tagRef, hasSpecTag := stream.Spec.Tags[tag]
		scheduled := false
		insecure := false
		importing := false

		var name string
		if hasSpecTag && tagRef.From != nil {
			if len(tagRef.From.Namespace) > 0 && tagRef.From.Namespace != stream.Namespace {
				name = fmt.Sprintf("%s/%s", tagRef.From.Namespace, tagRef.From.Name)
			} else {
				name = tagRef.From.Name
			}
			scheduled, insecure = tagRef.ImportPolicy.Scheduled, tagRef.ImportPolicy.Insecure
			gen := imageapi.LatestObservedTagGeneration(stream, tag)
			importing = !tagRef.Reference && tagRef.Generation != nil && *tagRef.Generation != gen
		}

		//   updates whenever tag :5.2 is changed

		// :latest (30 minutes ago) -> 102.205.358.453/foo/bar@sha256:abcde734
		//   error: last import failed 20 minutes ago
		//   updates automatically from index.docker.io/mysql/bar
		//     will use insecure HTTPS connections or HTTP
		//
		//   MySQL 5.5
		//   ---------
		//   Describes a system for updating based on practical changes to a database system
		//   with some other data involved
		//
		//   20 minutes ago  <import failed>
		//	  	Failed to locate the server in time
		//   30 minutes ago  102.205.358.453/foo/bar@sha256:abcdef
		//   1 hour ago      102.205.358.453/foo/bar@sha256:bfedfc

		//var shortErrors []string
		/*
			var internalReference *imageapi.DockerImageReference
			if value := stream.Status.DockerImageRepository; len(value) > 0 {
				ref, err := imageapi.ParseDockerImageReference(value)
				if err != nil {
					internalReference = &ref
				}
			}
		*/

		if referentialTags[tag].Len() > 0 {
			references := referentialTags[tag].List()
			imageapi.PrioritizeTags(references)
			fmt.Fprintf(out, "%s (%s)\n", tag, strings.Join(references, ", "))
		} else {
			fmt.Fprintf(out, "%s\n", tag)
		}

		switch {
		case !hasSpecTag || tagRef.From == nil:
			fmt.Fprintf(out, "  pushed image\n")
		case tagRef.From.Kind == "ImageStreamTag":
			switch {
			case tagRef.Reference:
				fmt.Fprintf(out, "  reference to %s\n", name)
			case scheduled:
				fmt.Fprintf(out, "  updates automatically from %s\n", name)
			default:
				fmt.Fprintf(out, "  tagged from %s\n", name)
			}
		case tagRef.From.Kind == "DockerImage":
			switch {
			case tagRef.Reference:
				fmt.Fprintf(out, "  reference to registry %s\n", name)
			case scheduled:
				fmt.Fprintf(out, "  updates automatically from registry %s\n", name)
			default:
				fmt.Fprintf(out, "  tagged from %s\n", name)
			}
		case tagRef.From.Kind == "ImageStreamImage":
			switch {
			case tagRef.Reference:
				fmt.Fprintf(out, "  reference to image %s\n", name)
			default:
				fmt.Fprintf(out, "  tagged from %s\n", name)
			}
		default:
			switch {
			case tagRef.Reference:
				fmt.Fprintf(out, "  reference to %s %s\n", tagRef.From.Kind, name)
			default:
				fmt.Fprintf(out, "  updates from %s %s\n", tagRef.From.Kind, name)
			}
		}
		if insecure {
			fmt.Fprintf(out, "    will use insecure HTTPS or HTTP connections\n")
		}

		fmt.Fprintln(out)

		extraOutput := false
		if d := tagRef.Annotations["description"]; len(d) > 0 {
			fmt.Fprintf(out, "  %s\n", d)
			extraOutput = true
		}
		if t := tagRef.Annotations["tags"]; len(t) > 0 {
			fmt.Fprintf(out, "  Tags: %s\n", strings.Join(strings.Split(t, ","), ", "))
			extraOutput = true
		}
		if t := tagRef.Annotations["supports"]; len(t) > 0 {
			fmt.Fprintf(out, "  Supports: %s\n", strings.Join(strings.Split(t, ","), ", "))
			extraOutput = true
		}
		if t := tagRef.Annotations["sampleRepo"]; len(t) > 0 {
			fmt.Fprintf(out, "  Example Repo: %s\n", t)
			extraOutput = true
		}
		if extraOutput {
			fmt.Fprintln(out)
		}

		if importing {
			fmt.Fprintf(out, "  ~ importing latest image ...\n")
		}

		for i := range taglist.Conditions {
			condition := &taglist.Conditions[i]
			switch condition.Type {
			case imageapi.ImportSuccess:
				if condition.Status == api.ConditionFalse {
					d := now.Sub(condition.LastTransitionTime.Time)
					fmt.Fprintf(out, "  ! error: Import failed (%s): %s\n      %s ago\n", condition.Reason, condition.Message, units.HumanDuration(d))
				}
			}
		}

		if len(taglist.Items) == 0 {
			continue
		}

		for i, event := range taglist.Items {
			d := now.Sub(event.Created.Time)

			if i == 0 {
				fmt.Fprintf(out, "  * %s\n", event.DockerImageReference)
			} else {
				fmt.Fprintf(out, "    %s\n", event.DockerImageReference)
			}

			ref, err := imageapi.ParseDockerImageReference(event.DockerImageReference)
			id := event.Image
			if len(id) > 0 && err == nil && ref.ID != id {
				fmt.Fprintf(out, "      %s ago\t%s\n", units.HumanDuration(d), id)
			} else {
				fmt.Fprintf(out, "      %s ago\n", units.HumanDuration(d))
			}
		}
	}
}
