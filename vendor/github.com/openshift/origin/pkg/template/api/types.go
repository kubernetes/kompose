package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"
)

// +genclient=true

// Template contains the inputs needed to produce a Config.
type Template struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// message is an optional instructional message that will
	// be displayed when this template is instantiated.
	// This field should inform the user how to utilize the newly created resources.
	// Parameter substitution will be performed on the message before being
	// displayed so that generated credentials and other parameters can be
	// included in the output.
	Message string

	// parameters is an optional array of Parameters used during the
	// Template to Config transformation.
	Parameters []Parameter

	// objects is an array of resources to include in this template.
	Objects []runtime.Object

	// objectLabels is an optional set of labels that are applied to every
	// object during the Template to Config transformation.
	ObjectLabels map[string]string
}

// TemplateList is a list of Template objects.
type TemplateList struct {
	unversioned.TypeMeta
	unversioned.ListMeta
	Items []Template
}

// Parameter defines a name/value variable that is to be processed during
// the Template to Config transformation.
type Parameter struct {
	// Required: Parameter name must be set and it can be referenced in Template
	// Items using ${PARAMETER_NAME}
	Name string

	// Optional: The name that will show in UI instead of parameter 'Name'
	DisplayName string

	// Optional: Parameter can have description
	Description string

	// Optional: Value holds the Parameter data. If specified, the generator
	// will be ignored. The value replaces all occurrences of the Parameter
	// ${Name} expression during the Template to Config transformation.
	Value string

	// Optional: Generate specifies the generator to be used to generate
	// random string from an input value specified by From field. The result
	// string is stored into Value field. If empty, no generator is being
	// used, leaving the result Value untouched.
	Generate string

	// Optional: From is an input value for the generator.
	From string

	// Optional: Indicates the parameter must have a value.  Defaults to false.
	Required bool
}

// These constants represent annotations keys affixed to templates
const (
	// TemplateDisplayName is an optional annotation that stores the name displayed by a UI when referencing a template.
	TemplateDisplayName = "openshift.io/display-name"
)
