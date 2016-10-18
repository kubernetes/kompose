package api

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/validation"
	"k8s.io/kubernetes/pkg/auth/user"
	"k8s.io/kubernetes/pkg/serviceaccount"
	"k8s.io/kubernetes/pkg/util/sets"
	// uservalidation "github.com/openshift/origin/pkg/user/api/validation"
)

// NormalizeResources expands all resource groups and forces all resources to lower case.
// If the rawResources are already normalized, it returns the original set to avoid the
// allocation and GC cost, since this is hit multiple times for every REST call.
// That means you should NEVER MODIFY THE RESULT of this call.
func NormalizeResources(rawResources sets.String) sets.String {
	// we only need to expand groups if the exist and we don't create them with groups
	// by default.  Only accept the cost of expansion if we're doing work.
	needsNormalization := false
	for currResource := range rawResources {
		if needsNormalizing(currResource) {
			needsNormalization = true
			break
		}

	}
	if !needsNormalization {
		return rawResources
	}

	ret := sets.String{}
	toVisit := rawResources.List()
	visited := sets.String{}

	for i := 0; i < len(toVisit); i++ {
		currResource := toVisit[i]
		if visited.Has(currResource) {
			continue
		}
		visited.Insert(currResource)

		if !strings.HasPrefix(currResource, resourceGroupPrefix) {
			ret.Insert(strings.ToLower(currResource))
			continue
		}

		if resourceTypes, exists := groupsToResources[currResource]; exists {
			toVisit = append(toVisit, resourceTypes...)
		}
	}

	return ret
}

func needsNormalizing(in string) bool {
	if strings.HasPrefix(in, resourceGroupPrefix) {
		return true
	}
	for _, r := range in {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func (r PolicyRule) String() string {
	return "PolicyRule" + r.CompactString()
}

// CompactString exposes a compact string representation for use in escalation error messages
func (r PolicyRule) CompactString() string {
	formatStringParts := []string{}
	formatArgs := []interface{}{}
	if len(r.Verbs) > 0 {
		formatStringParts = append(formatStringParts, "Verbs:%q")
		formatArgs = append(formatArgs, r.Verbs.List())
	}
	if len(r.APIGroups) > 0 {
		formatStringParts = append(formatStringParts, "APIGroups:%q")
		formatArgs = append(formatArgs, r.APIGroups)
	}
	if len(r.Resources) > 0 {
		formatStringParts = append(formatStringParts, "Resources:%q")
		formatArgs = append(formatArgs, r.Resources.List())
	}
	if len(r.ResourceNames) > 0 {
		formatStringParts = append(formatStringParts, "ResourceNames:%q")
		formatArgs = append(formatArgs, r.ResourceNames.List())
	}
	if r.AttributeRestrictions != nil {
		formatStringParts = append(formatStringParts, "Restrictions:%q")
		formatArgs = append(formatArgs, r.AttributeRestrictions)
	}
	if len(r.NonResourceURLs) > 0 {
		formatStringParts = append(formatStringParts, "NonResourceURLs:%q")
		formatArgs = append(formatArgs, r.NonResourceURLs.List())
	}
	formatString := "{" + strings.Join(formatStringParts, ", ") + "}"
	return fmt.Sprintf(formatString, formatArgs...)
}

func getRoleBindingValues(roleBindingMap map[string]*RoleBinding) []*RoleBinding {
	ret := []*RoleBinding{}
	for _, currBinding := range roleBindingMap {
		ret = append(ret, currBinding)
	}

	return ret
}
func SortRoleBindings(roleBindingMap map[string]*RoleBinding, reverse bool) []*RoleBinding {
	roleBindings := getRoleBindingValues(roleBindingMap)
	if reverse {
		sort.Sort(sort.Reverse(RoleBindingSorter(roleBindings)))
	} else {
		sort.Sort(RoleBindingSorter(roleBindings))
	}

	return roleBindings
}

type PolicyBindingSorter []PolicyBinding

func (s PolicyBindingSorter) Len() int {
	return len(s)
}
func (s PolicyBindingSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s PolicyBindingSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type RoleBindingSorter []*RoleBinding

func (s RoleBindingSorter) Len() int {
	return len(s)
}
func (s RoleBindingSorter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}
func (s RoleBindingSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func GetPolicyBindingName(policyRefNamespace string) string {
	return fmt.Sprintf("%s:%s", policyRefNamespace, PolicyName)
}

var ClusterPolicyBindingName = GetPolicyBindingName("")

func BuildSubjects(users, groups []string, userNameValidator, groupNameValidator validation.ValidateNameFunc) []kapi.ObjectReference {
	subjects := []kapi.ObjectReference{}

	for _, user := range users {
		saNamespace, saName, err := serviceaccount.SplitUsername(user)
		if err == nil {
			subjects = append(subjects, kapi.ObjectReference{Kind: ServiceAccountKind, Namespace: saNamespace, Name: saName})
			continue
		}

		kind := UserKind
		if len(userNameValidator(user, false)) != 0 {
			kind = SystemUserKind
		}

		subjects = append(subjects, kapi.ObjectReference{Kind: kind, Name: user})
	}

	for _, group := range groups {
		kind := GroupKind
		if len(groupNameValidator(group, false)) != 0 {
			kind = SystemGroupKind
		}

		subjects = append(subjects, kapi.ObjectReference{Kind: kind, Name: group})
	}

	return subjects
}

// StringSubjectsFor returns users and groups for comparison against user.Info.  currentNamespace is used to
// to create usernames for service accounts where namespace=="".
func StringSubjectsFor(currentNamespace string, subjects []kapi.ObjectReference) ([]string, []string) {
	// these MUST be nil to indicate empty
	var users, groups []string

	for _, subject := range subjects {
		switch subject.Kind {
		case ServiceAccountKind:
			namespace := currentNamespace
			if len(subject.Namespace) > 0 {
				namespace = subject.Namespace
			}
			if len(namespace) > 0 {
				users = append(users, serviceaccount.MakeUsername(namespace, subject.Name))
			}

		case UserKind, SystemUserKind:
			users = append(users, subject.Name)

		case GroupKind, SystemGroupKind:
			groups = append(groups, subject.Name)
		}
	}

	return users, groups
}

// SubjectsStrings returns users, groups, serviceaccounts, unknown for display purposes.  currentNamespace is used to
// hide the subject.Namespace for ServiceAccounts in the currentNamespace
func SubjectsStrings(currentNamespace string, subjects []kapi.ObjectReference) ([]string, []string, []string, []string) {
	users := []string{}
	groups := []string{}
	sas := []string{}
	others := []string{}

	for _, subject := range subjects {
		switch subject.Kind {
		case ServiceAccountKind:
			if len(subject.Namespace) > 0 && currentNamespace != subject.Namespace {
				sas = append(sas, subject.Namespace+"/"+subject.Name)
			} else {
				sas = append(sas, subject.Name)
			}

		case UserKind, SystemUserKind:
			users = append(users, subject.Name)

		case GroupKind, SystemGroupKind:
			groups = append(groups, subject.Name)

		default:
			others = append(others, fmt.Sprintf("%s/%s/%s", subject.Kind, subject.Namespace, subject.Name))

		}
	}

	return users, groups, sas, others
}

// SubjectsContainUser returns true if the provided subjects contain the named user. currentNamespace
// is used to identify service accounts that are defined in a relative fashion.
func SubjectsContainUser(subjects []kapi.ObjectReference, currentNamespace string, user string) bool {
	if !strings.HasPrefix(user, serviceaccount.ServiceAccountUsernamePrefix) {
		for _, subject := range subjects {
			switch subject.Kind {
			case UserKind, SystemUserKind:
				if user == subject.Name {
					return true
				}
			}
		}
		return false
	}

	for _, subject := range subjects {
		switch subject.Kind {
		case ServiceAccountKind:
			namespace := currentNamespace
			if len(subject.Namespace) > 0 {
				namespace = subject.Namespace
			}
			if len(namespace) == 0 {
				continue
			}
			if user == serviceaccount.MakeUsername(namespace, subject.Name) {
				return true
			}

		case UserKind, SystemUserKind:
			if user == subject.Name {
				return true
			}
		}
	}
	return false
}

// SubjectsContainAnyGroup returns true if the provided subjects any of the named groups.
func SubjectsContainAnyGroup(subjects []kapi.ObjectReference, groups []string) bool {
	for _, subject := range subjects {
		switch subject.Kind {
		case GroupKind, SystemGroupKind:
			for _, group := range groups {
				if group == subject.Name {
					return true
				}
			}
		}
	}
	return false
}

func AddUserToSAR(user user.Info, sar *SubjectAccessReview) *SubjectAccessReview {
	origScopes := user.GetExtra()[ScopesKey]
	scopes := make([]string, len(origScopes), len(origScopes))
	copy(scopes, origScopes)

	sar.User = user.GetName()
	sar.Groups = sets.NewString(user.GetGroups()...)
	sar.Scopes = scopes
	return sar
}
func AddUserToLSAR(user user.Info, lsar *LocalSubjectAccessReview) *LocalSubjectAccessReview {
	origScopes := user.GetExtra()[ScopesKey]
	scopes := make([]string, len(origScopes), len(origScopes))
	copy(scopes, origScopes)

	lsar.User = user.GetName()
	lsar.Groups = sets.NewString(user.GetGroups()...)
	lsar.Scopes = scopes
	return lsar
}

// +gencopy=false
// PolicyRuleBuilder let's us attach methods.  A no-no for API types
type PolicyRuleBuilder struct {
	PolicyRule PolicyRule
}

func NewRule(verbs ...string) *PolicyRuleBuilder {
	return &PolicyRuleBuilder{
		PolicyRule: PolicyRule{
			Verbs:         sets.NewString(verbs...),
			Resources:     sets.String{},
			ResourceNames: sets.String{},
		},
	}
}

func (r *PolicyRuleBuilder) Groups(groups ...string) *PolicyRuleBuilder {
	r.PolicyRule.APIGroups = append(r.PolicyRule.APIGroups, groups...)
	return r
}

func (r *PolicyRuleBuilder) Resources(resources ...string) *PolicyRuleBuilder {
	r.PolicyRule.Resources.Insert(resources...)
	return r
}

func (r *PolicyRuleBuilder) Names(names ...string) *PolicyRuleBuilder {
	r.PolicyRule.ResourceNames.Insert(names...)
	return r
}

func (r *PolicyRuleBuilder) RuleOrDie() PolicyRule {
	ret, err := r.Rule()
	if err != nil {
		panic(err)
	}
	return ret
}

func (r *PolicyRuleBuilder) Rule() (PolicyRule, error) {
	if len(r.PolicyRule.Verbs) == 0 {
		return PolicyRule{}, fmt.Errorf("verbs are required: %#v", r.PolicyRule)
	}

	switch {
	case len(r.PolicyRule.NonResourceURLs) > 0:
		if len(r.PolicyRule.APIGroups) != 0 || len(r.PolicyRule.Resources) != 0 || len(r.PolicyRule.ResourceNames) != 0 {
			return PolicyRule{}, fmt.Errorf("non-resource rule may not have apiGroups, resources, or resourceNames: %#v", r.PolicyRule)
		}
	case len(r.PolicyRule.Resources) > 0:
		if len(r.PolicyRule.NonResourceURLs) != 0 {
			return PolicyRule{}, fmt.Errorf("resource rule may not have nonResourceURLs: %#v", r.PolicyRule)
		}
		if len(r.PolicyRule.APIGroups) == 0 {
			return PolicyRule{}, fmt.Errorf("resource rule must have apiGroups: %#v", r.PolicyRule)
		}
	default:
		return PolicyRule{}, fmt.Errorf("a rule must have either nonResourceURLs or resources: %#v", r.PolicyRule)
	}

	return r.PolicyRule, nil
}

type SortableRuleSlice []PolicyRule

func (s SortableRuleSlice) Len() int      { return len(s) }
func (s SortableRuleSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortableRuleSlice) Less(i, j int) bool {
	return strings.Compare(s[i].String(), s[j].String()) < 0
}
