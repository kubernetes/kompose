package clientcmd

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

// negotiateVersion queries the server's supported api versions to find a version that both client and server support.
// - If no version is provided, try registered client versions in order of preference.
// - If version is provided, but not default config (explicitly requested via
//   commandline flag), and is unsupported by the server, print a warning to
//   stderr and try client's registered versions in order of preference.
// - If version is config default, and the server does not support it, return an error.
func negotiateVersion(client *kclient.Client, config *restclient.Config, requestedGV *unversioned.GroupVersion, clientGVs []unversioned.GroupVersion) (*unversioned.GroupVersion, error) {
	// Ensure we have a client
	var err error
	if client == nil {
		client, err = kclient.New(config)
		if err != nil {
			return nil, err
		}
	}

	// Determine our preferred version
	preferredGV := copyGroupVersion(requestedGV)
	if preferredGV == nil {
		preferredGV = copyGroupVersion(config.GroupVersion)
	}

	// Get server versions
	serverGVs, err := serverAPIVersions(client, "/oapi")
	if err != nil {
		if errors.IsNotFound(err) {
			glog.V(4).Infof("Server path /oapi was not found, returning the requested group version %v", preferredGV)
			return preferredGV, nil
		}
		return nil, err
	}

	// Find a version we can all agree on
	matchedGV, err := matchAPIVersion(preferredGV, clientGVs, serverGVs)
	if err != nil {
		return nil, err
	}

	// Enforce a match if the preferredGV is the config default
	if config.GroupVersion != nil && (*preferredGV == *config.GroupVersion) && (*matchedGV != *config.GroupVersion) {
		return nil, fmt.Errorf("server does not support API version %q", config.GroupVersion.String())
	}

	return matchedGV, err
}

// serverAPIVersions fetches the server versions available from the groupless API at the given prefix
func serverAPIVersions(c *kclient.Client, grouplessPrefix string) ([]unversioned.GroupVersion, error) {
	// Get versions doc
	var v unversioned.APIVersions
	if err := c.Get().AbsPath(grouplessPrefix).Do().Into(&v); err != nil {
		return []unversioned.GroupVersion{}, err
	}

	// Convert to GroupVersion structs
	serverAPIVersions := []unversioned.GroupVersion{}
	for _, version := range v.Versions {
		gv, err := unversioned.ParseGroupVersion(version)
		if err != nil {
			return []unversioned.GroupVersion{}, err
		}
		serverAPIVersions = append(serverAPIVersions, gv)
	}
	return serverAPIVersions, nil
}

func matchAPIVersion(preferredGV *unversioned.GroupVersion, clientGVs []unversioned.GroupVersion, serverGVs []unversioned.GroupVersion) (*unversioned.GroupVersion, error) {
	// If version explicitly requested verify that both client and server support it.
	// If server does not support warn, but try to negotiate a lower version.
	if preferredGV != nil {
		if !containsGroupVersion(clientGVs, *preferredGV) {
			return nil, fmt.Errorf("client does not support API version %q; client supported API versions: %v", preferredGV, clientGVs)
		}
		if containsGroupVersion(serverGVs, *preferredGV) {
			return preferredGV, nil
		}
	}

	for _, clientGV := range clientGVs {
		if containsGroupVersion(serverGVs, clientGV) {
			t := clientGV
			return &t, nil
		}
	}
	return nil, fmt.Errorf("failed to negotiate an api version; server supports: %v, client supports: %v", serverGVs, clientGVs)
}

func copyGroupVersion(version *unversioned.GroupVersion) *unversioned.GroupVersion {
	if version == nil {
		return nil
	}
	versionCopy := *version
	return &versionCopy
}

func containsGroupVersion(versions []unversioned.GroupVersion, version unversioned.GroupVersion) bool {
	for _, v := range versions {
		if v == version {
			return true
		}
	}
	return false
}
