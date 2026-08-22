package policy

import (
	"fmt"
	"sort"
	"strings"

	"licensecompat.local/internal/model"
)

type Verdict struct {
	RuleID      string
	Kind        model.FindingKind
	Code        string
	Message     string
	Licenses    []string
	AllowWaiver bool
}

func Evaluate(policy model.Policy, licenses []string) []Verdict {
	licenses = canonical(licenses)
	verdicts := make([]Verdict, 0)
	for _, rule := range policy.Rules {
		for _, forbidden := range canonical(rule.Forbidden) {
			if contains(licenses, forbidden) {
				verdicts = append(verdicts, Verdict{RuleID: rule.ID, Kind: model.FindingBlocker, Code: "policy.forbidden." + strings.ToLower(forbidden), Message: fmt.Sprintf("policy %q forbids license %s", rule.Name, forbidden), Licenses: []string{forbidden}, AllowWaiver: rule.AllowWaiver})
			}
		}
		for _, required := range canonical(rule.RequireNote) {
			if contains(licenses, required) {
				verdicts = append(verdicts, Verdict{RuleID: rule.ID, Kind: model.FindingObligation, Code: "policy.notice." + strings.ToLower(required), Message: fmt.Sprintf("policy %q requires a notice for %s", rule.Name, required), Licenses: []string{required}, AllowWaiver: false})
			}
		}
	}
	return unique(verdicts)
}

func PairCompatibility(left, right string) (model.FindingKind, string, bool) {
	l, r := Lookup(left), Lookup(right)
	if l.ID == "" || r.ID == "" {
		return model.FindingWarning, "license declaration is incomplete", false
	}
	if (l.Family == FamilyStrongCopy && r.Family == FamilyProprietary) || (r.Family == FamilyStrongCopy && l.Family == FamilyProprietary) {
		return model.FindingBlocker, "strong copyleft dependency conflicts with proprietary delivery", true
	}
	if l.Family == FamilyUnknown || r.Family == FamilyUnknown {
		return model.FindingWarning, "license family is unknown and needs review", false
	}
	if l.Source || r.Source {
		return model.FindingObligation, "dependency introduces source availability obligations", false
	}
	return model.FindingObligation, "license notice must be retained", false
}

func canonical(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = Normalize(value)
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
func unique(values []Verdict) []Verdict {
	seen := map[string]bool{}
	out := make([]Verdict, 0, len(values))
	for _, v := range values {
		key := v.RuleID + ":" + v.Code
		if !seen[key] {
			seen[key] = true
			out = append(out, v)
		}
	}
	return out
}
