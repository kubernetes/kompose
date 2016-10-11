package config

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/util/homedir"
)

const (
	OpenShiftConfigPathEnvVar      = "KUBECONFIG"
	OpenShiftConfigFlagName        = "config"
	OpenShiftConfigHomeDir         = ".kube"
	OpenShiftConfigHomeFileName    = "config"
	OpenShiftConfigHomeDirFileName = OpenShiftConfigHomeDir + "/" + OpenShiftConfigHomeFileName
)

var RecommendedHomeFile = path.Join(homedir.HomeDir(), OpenShiftConfigHomeDirFileName)

// currentMigrationRules returns a map that holds the history of recommended home directories used in previous versions.
// Any future changes to RecommendedHomeFile and related are expected to add a migration rule here, in order to make
// sure existing config files are migrated to their new locations properly.
func currentMigrationRules() map[string]string {
	oldRecommendedHomeFile := path.Join(homedir.HomeDir(), ".kube/.config")
	oldRecommendedWindowsHomeFile := path.Join(os.Getenv("HOME"), OpenShiftConfigHomeDirFileName)

	migrationRules := map[string]string{}
	migrationRules[RecommendedHomeFile] = oldRecommendedHomeFile
	if runtime.GOOS == "windows" {
		migrationRules[RecommendedHomeFile] = oldRecommendedWindowsHomeFile
	}
	return migrationRules
}

// NewOpenShiftClientConfigLoadingRules returns file priority loading rules for OpenShift.
// 1. --config value
// 2. if KUBECONFIG env var has a value, use it. Otherwise, ~/.kube/config file
func NewOpenShiftClientConfigLoadingRules() *clientcmd.ClientConfigLoadingRules {
	chain := []string{}

	envVarFile := os.Getenv(OpenShiftConfigPathEnvVar)
	if len(envVarFile) != 0 {
		chain = append(chain, filepath.SplitList(envVarFile)...)
	} else {
		chain = append(chain, RecommendedHomeFile)
	}

	return &clientcmd.ClientConfigLoadingRules{
		Precedence:     chain,
		MigrationRules: currentMigrationRules(),
	}
}

func NewPathOptions(cmd *cobra.Command) *kclientcmd.PathOptions {
	return NewPathOptionsWithConfig(kcmdutil.GetFlagString(cmd, OpenShiftConfigFlagName))
}

func NewPathOptionsWithConfig(configPath string) *kclientcmd.PathOptions {
	return &kclientcmd.PathOptions{
		GlobalFile: RecommendedHomeFile,

		EnvVar:           OpenShiftConfigPathEnvVar,
		ExplicitFileFlag: OpenShiftConfigFlagName,

		LoadingRules: &kclientcmd.ClientConfigLoadingRules{
			ExplicitPath: configPath,
		},
	}
}
