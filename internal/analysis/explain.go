package analysis

import (
	"sort"

	"licensecompat.local/internal/model"
)

type ExplainStep struct {
	Subject  string   `json:"subject"`
	Evidence string   `json:"evidence"`
	Parents  []string `json:"parents,omitempty"`
}
type Explanation struct {
	Finding model.Finding `json:"finding"`
	Steps   []ExplainStep `json:"steps"`
}

func Explain(finding model.Finding, submission model.Submission) Explanation {
	components := map[string]model.Component{}
	for _, component := range submission.Components {
		components[component.ID] = component
	}
	steps := make([]ExplainStep, 0, len(finding.Components)+len(finding.Derivation))
	for _, id := range finding.Components {
		component := components[id]
		steps = append(steps, ExplainStep{Subject: id, Evidence: component.Name + "@" + component.Version + " declares " + component.License})
	}
	for _, item := range finding.Derivation {
		steps = append(steps, ExplainStep{Subject: item, Evidence: "derived during deterministic analysis"})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].Subject < steps[j].Subject })
	return Explanation{Finding: finding, Steps: steps}
}
