package model

import (
	"fmt"
	"sort"
	"strings"
)

func ValidatePolicy(p Policy) error {
	if strings.TrimSpace(p.Name) == "" {
		return Invalid("name", "must not be empty")
	}
	if p.Version < 1 {
		return Invalid("version", "must be positive")
	}
	if p.Status != "" && p.Status != PolicyDraft && p.Status != PolicyActive && p.Status != PolicyRetired {
		return Invalid("status", "unknown policy status")
	}
	seen := map[string]bool{}
	for i, r := range p.Rules {
		if strings.TrimSpace(r.ID) == "" {
			return Invalid(fmt.Sprintf("rules[%d].id", i), "must not be empty")
		}
		if seen[r.ID] {
			return Invalid(fmt.Sprintf("rules[%d].id", i), "must be unique")
		}
		seen[r.ID] = true
		if strings.TrimSpace(r.Name) == "" {
			return Invalid(fmt.Sprintf("rules[%d].name", i), "must not be empty")
		}
		if len(r.Forbidden) == 0 && len(r.RequireNote) == 0 {
			return Invalid(fmt.Sprintf("rules[%d]", i), "must define a consequence")
		}
	}
	return nil
}

func ValidateSubmission(s Submission) error {
	if strings.TrimSpace(s.Name) == "" {
		return Invalid("name", "must not be empty")
	}
	if strings.TrimSpace(s.Release) == "" {
		return Invalid("release", "must not be empty")
	}
	if len(s.Components) == 0 {
		return Invalid("components", "must contain at least one component")
	}
	seen := map[string]bool{}
	for i, c := range s.Components {
		if strings.TrimSpace(c.ID) == "" {
			return Invalid(fmt.Sprintf("components[%d].id", i), "must not be empty")
		}
		if seen[c.ID] {
			return Invalid(fmt.Sprintf("components[%d].id", i), "must be unique")
		}
		seen[c.ID] = true
		if strings.TrimSpace(c.Name) == "" {
			return Invalid(fmt.Sprintf("components[%d].name", i), "must not be empty")
		}
		if strings.TrimSpace(c.Version) == "" {
			return Invalid(fmt.Sprintf("components[%d].version", i), "must not be empty")
		}
		if c.Status != "" && c.Status != ComponentPending && c.Status != ComponentTrusted && c.Status != ComponentDisputed && c.Status != ComponentRevoked {
			return Invalid(fmt.Sprintf("components[%d].status", i), "unknown component status")
		}
	}
	for i, e := range s.Edges {
		if !seen[e.From] || !seen[e.To] {
			return Invalid(fmt.Sprintf("edges[%d]", i), "references unknown component")
		}
		if e.From == e.To {
			return Invalid(fmt.Sprintf("edges[%d]", i), "self dependency is not allowed")
		}
		if e.Scope == "" {
			return Invalid(fmt.Sprintf("edges[%d].scope", i), "must not be empty")
		}
	}
	return nil
}

func CanonicalLicenses(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
		if value != "" && !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

func CanTransitionAnalysis(from, to AnalysisStatus) bool {
	allowed := map[AnalysisStatus][]AnalysisStatus{
		AnalysisQueued:     {AnalysisRunning, AnalysisSuperseded},
		AnalysisRunning:    {AnalysisBlocked, AnalysisReviewable, AnalysisQueued},
		AnalysisBlocked:    {AnalysisReviewable, AnalysisPublished, AnalysisSuperseded},
		AnalysisReviewable: {AnalysisBlocked, AnalysisPublished, AnalysisSuperseded},
		AnalysisPublished:  {AnalysisSuperseded},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return true
		}
	}
	return false
}

func CanTransitionWaiver(from, to WaiverStatus) bool {
	return (from == WaiverRequested && (to == WaiverApproved || to == WaiverRejected)) || (from == WaiverApproved && to == WaiverExpired)
}

func CanWaiveFinding(kind FindingKind) bool { return kind == FindingBlocker }
