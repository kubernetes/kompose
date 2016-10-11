package util

import (
	"path/filepath"
)

const (
	ProductOrigin           = `Origin`
	ProductOpenShift        = `OpenShift`
	ProductAtomicEnterprise = `Atomic Enterprise`
)

// GetProductName chooses appropriate product for a binary name.
func GetProductName(binaryName string) string {
	name := filepath.Base(binaryName)
	for {
		switch name {
		case "openshift":
			return ProductOpenShift
		case "atomic-enterprise":
			return ProductAtomicEnterprise
		default:
			return ProductOrigin
		}
	}
}

// GetPlatformName returns an appropriate platform name for given binary name.
// Platform name can be used as a headline in command's usage.
func GetPlatformName(binaryName string) string {
	switch GetProductName(binaryName) {
	case ProductAtomicEnterprise:
		return "Atomic Enterprise Platform"
	case ProductOpenShift:
		return "OpenShift Application Platform"
	default:
		return "Origin Application Platform"
	}
}

// GetDistributionName returns an appropriate Kubernetes distribution name.
// Distribution name can be used in relation to some feature set in command's
// usage string (e.g. <distribution name> allows you to build, run, etc.).
func GetDistributionName(binaryName string) string {
	switch GetProductName(binaryName) {
	case ProductAtomicEnterprise:
		return "Atomic distribution of Kubernetes"
	case ProductOpenShift:
		return "OpenShift distribution of Kubernetes"
	default:
		return "Origin distribution of Kubernetes"
	}

}
