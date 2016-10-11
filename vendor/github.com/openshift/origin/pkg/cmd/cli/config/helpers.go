package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/openshift/origin/pkg/cmd/util"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

// TODO should be moved upstream
func RelativizeClientConfigPaths(cfg *clientcmdapi.Config, base string) (err error) {
	for k, cluster := range cfg.Clusters {
		if len(cluster.CertificateAuthority) > 0 {
			if cluster.CertificateAuthority, err = util.MakeAbs(cluster.CertificateAuthority, ""); err != nil {
				return err
			}
			if cluster.CertificateAuthority, err = util.MakeRelative(cluster.CertificateAuthority, base); err != nil {
				return err
			}
			cfg.Clusters[k] = cluster
		}
	}
	for k, authInfo := range cfg.AuthInfos {
		if len(authInfo.ClientCertificate) > 0 {
			if authInfo.ClientCertificate, err = util.MakeAbs(authInfo.ClientCertificate, ""); err != nil {
				return err
			}
			if authInfo.ClientCertificate, err = util.MakeRelative(authInfo.ClientCertificate, base); err != nil {
				return err
			}
		}
		if len(authInfo.ClientKey) > 0 {
			if authInfo.ClientKey, err = util.MakeAbs(authInfo.ClientKey, ""); err != nil {
				return err
			}
			if authInfo.ClientKey, err = util.MakeRelative(authInfo.ClientKey, base); err != nil {
				return err
			}
		}
		cfg.AuthInfos[k] = authInfo
	}
	return nil
}

var validURLSchemes = []string{"https://", "http://", "tcp://"}

// NormalizeServerURL is opinionated normalization of a string that represents a URL. Returns the URL provided matching the format
// expected when storing a URL in a config. Sets a scheme and port if not present, removes unnecessary trailing
// slashes, etc. Can be used to normalize a URL provided by user input.
func NormalizeServerURL(s string) (string, error) {
	// normalize scheme
	if !hasScheme(s) {
		s = validURLSchemes[0] + s
	}

	addr, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("Not a valid URL: %v.", err)
	}

	// normalize host:port
	if strings.Contains(addr.Host, ":") {
		_, port, err := net.SplitHostPort(addr.Host)
		if err != nil {
			return "", fmt.Errorf("Not a valid host:port: %v.", err)
		}
		_, err = strconv.ParseUint(port, 10, 16)
		if err != nil {
			return "", fmt.Errorf("Not a valid port: %v. Port numbers must be between 0 and 65535.", port)
		}
	} else {
		port := 0
		switch addr.Scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		default:
			return "", fmt.Errorf("No port specified.")
		}
		addr.Host = net.JoinHostPort(addr.Host, strconv.FormatInt(int64(port), 10))
	}

	// remove trailing slash if that's the only path we have
	if addr.Path == "/" {
		addr.Path = ""
	}

	return addr.String(), nil
}

func hasScheme(s string) bool {
	for _, p := range validURLSchemes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
