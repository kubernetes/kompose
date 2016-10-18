package clientcmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift/origin/pkg/cmd/cli/config"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
)

func DefaultClientConfig(flags *pflag.FlagSet) clientcmd.ClientConfig {
	loadingRules := config.NewOpenShiftClientConfigLoadingRules()
	flags.StringVar(&loadingRules.ExplicitPath, config.OpenShiftConfigFlagName, "", "Path to the config file to use for CLI requests.")
	cobra.MarkFlagFilename(flags, config.OpenShiftConfigFlagName)

	overrides := &clientcmd.ConfigOverrides{}
	overrideFlags := clientcmd.RecommendedConfigOverrideFlags("")
	overrideFlags.ContextOverrideFlags.Namespace.ShortName = "n"
	overrideFlags.AuthOverrideFlags.Username.LongName = ""
	overrideFlags.AuthOverrideFlags.Password.LongName = ""
	clientcmd.BindOverrideFlags(overrides, flags, overrideFlags)
	cobra.MarkFlagFilename(flags, overrideFlags.AuthOverrideFlags.ClientCertificate.LongName)
	cobra.MarkFlagFilename(flags, overrideFlags.AuthOverrideFlags.ClientKey.LongName)
	cobra.MarkFlagFilename(flags, overrideFlags.ClusterOverrideFlags.CertificateAuthority.LongName)

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	return clientConfig
}
