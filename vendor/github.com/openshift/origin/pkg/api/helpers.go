package api

import (
	"fmt"
	"strings"

	"k8s.io/kubernetes/pkg/api/validation"
)

var NameMayNotBe = []string{".", ".."}
var NameMayNotContain = []string{"/", "%"}

func MinimalNameRequirements(name string, prefix bool) []string {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return []string{fmt.Sprintf(`name may not be %q`, illegalName)}
		}
	}

	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			return []string{fmt.Sprintf(`name may not contain %q`, illegalContent)}
		}
	}

	return nil
}

// GetNameValidationFunc returns a name validation function that includes the standard restrictions we want for all types
func GetNameValidationFunc(nameFunc validation.ValidateNameFunc) validation.ValidateNameFunc {
	return func(name string, prefix bool) []string {
		if reasons := MinimalNameRequirements(name, prefix); len(reasons) != 0 {
			return reasons
		}

		return nameFunc(name, prefix)
	}
}

// GetFieldLabelConversionFunc returns a field label conversion func, which does the following:
// * returns overrideLabels[label], value, nil if the specified label exists in the overrideLabels map
// * returns label, value, nil if the specified label exists as a key in the supportedLabels map (values in this map are unused, it is intended to be a prototypical label/value map)
// * otherwise, returns an error
func GetFieldLabelConversionFunc(supportedLabels map[string]string, overrideLabels map[string]string) func(label, value string) (string, string, error) {
	return func(label, value string) (string, string, error) {
		if label, overridden := overrideLabels[label]; overridden {
			return label, value, nil
		}
		if _, supported := supportedLabels[label]; supported {
			return label, value, nil
		}
		return "", "", fmt.Errorf("field label not supported: %s", label)
	}
}
