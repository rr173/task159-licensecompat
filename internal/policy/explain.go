package policy

import (
	"fmt"
	"strings"

	"licensecompat.local/internal/model"
)

type Explanation struct {
	License string   `json:"license"`
	Family  Family   `json:"family"`
	Facts   []string `json:"facts"`
}

func ExplainLicense(value string) Explanation {
	info := Lookup(value)
	facts := []string{fmt.Sprintf("normalized license is %s", info.ID), fmt.Sprintf("license family is %s", info.Family)}
	if info.Notice {
		facts = append(facts, "retains a notice obligation")
	}
	if info.Source {
		facts = append(facts, "may require source availability")
	}
	if info.Patent {
		facts = append(facts, "contains patent-related terms")
	}
	if info.Family == FamilyUnknown {
		facts = append(facts, "is not in the built-in catalog and needs legal review")
	}
	return Explanation{License: info.ID, Family: info.Family, Facts: facts}
}

func DescribeFinding(f model.Finding) string {
	parts := []string{f.Message}
	if len(f.Components) > 0 {
		parts = append(parts, "components: "+strings.Join(f.Components, ", "))
	}
	if len(f.Derivation) > 0 {
		parts = append(parts, "derived by: "+strings.Join(f.Derivation, " -> "))
	}
	return strings.Join(parts, "; ")
}
