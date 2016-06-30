package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/homedir"
	"github.com/docker/docker/pkg/tlsconfig"
	dockerclient "github.com/fsouza/go-dockerclient"
)

const (
	// DefaultAPIVersion is the default docker API version set by libcompose
	DefaultAPIVersion   = "1.20"
	defaultTrustKeyFile = "key.json"
	defaultCaFile       = "ca.pem"
	defaultKeyFile      = "key.pem"
	defaultCertFile     = "cert.pem"
)

var (
	dockerCertPath = os.Getenv("DOCKER_CERT_PATH")
)

func init() {
	if dockerCertPath == "" {
		dockerCertPath = cliconfig.ConfigDir()
	}
}

// ClientOpts holds docker client options (host, tls, ..)
type ClientOpts struct {
	TLS        bool
	TLSVerify  bool
	TLSOptions tlsconfig.Options
	TrustKey   string
	Host       string
	APIVersion string
}

// CreateClient creates a docker client based on the specified options.
func CreateClient(c ClientOpts) (*dockerclient.Client, error) {
	if c.TLSOptions.CAFile == "" {
		c.TLSOptions.CAFile = filepath.Join(dockerCertPath, defaultCaFile)
	}
	if c.TLSOptions.CertFile == "" {
		c.TLSOptions.CertFile = filepath.Join(dockerCertPath, defaultCertFile)
	}
	if c.TLSOptions.KeyFile == "" {
		c.TLSOptions.KeyFile = filepath.Join(dockerCertPath, defaultKeyFile)
	}

	if c.Host == "" {
		defaultHost := os.Getenv("DOCKER_HOST")
		if defaultHost == "" {
			if runtime.GOOS != "windows" {
				// If we do not have a host, default to unix socket
				defaultHost = fmt.Sprintf("unix://%s", opts.DefaultUnixSocket)
			} else {
				// If we do not have a host, default to TCP socket on Windows
				defaultHost = fmt.Sprintf("tcp://%s:%d", opts.DefaultHTTPHost, opts.DefaultHTTPPort)
			}
		}
		defaultHost, err := opts.ValidateHost(defaultHost)
		if err != nil {
			return nil, err
		}
		c.Host = defaultHost
	}

	if c.TrustKey == "" {
		c.TrustKey = filepath.Join(homedir.Get(), ".docker", defaultTrustKeyFile)
	}

	if c.TLSVerify {
		c.TLS = true
	}

	if c.TLS {
		c.TLSOptions.InsecureSkipVerify = !c.TLSVerify
	}

	apiVersion := c.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}
	if c.TLS {
		client, err := dockerclient.NewVersionedTLSClient(c.Host, c.TLSOptions.CertFile, c.TLSOptions.KeyFile, c.TLSOptions.CAFile, apiVersion)
		if err != nil {
			return nil, err
		}
		if c.TLSOptions.InsecureSkipVerify {
			client.TLSConfig.InsecureSkipVerify = true
		}
		return client, nil
	}
	return dockerclient.NewVersionedClient(c.Host, apiVersion)
}
