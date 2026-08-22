package analysis

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"licensecompat.local/internal/model"
	"licensecompat.local/internal/policy"
)

type Result struct {
	Findings []model.Finding
	Ordered  []string
}

func Evaluate(submission model.Submission, active model.Policy, analysisID string, now time.Time) (Result, error) {
	graph, err := BuildGraph(submission)
	if err != nil {
		return Result{}, err
	}
	ordered, err := graph.Topological()
	if err != nil {
		return Result{}, err
	}
	findings := make([]model.Finding, 0)
	licenses := make([]string, 0, len(ordered))
	for _, id := range ordered {
		c, _ := graph.Component(id)
		licenses = append(licenses, c.License)
		findings = append(findings, componentFindings(c, analysisID, now)...)
	}
	for _, verdict := range policy.Evaluate(active, licenses) {
		components := componentsForLicenses(graph, verdict.Licenses)
		findings = append(findings, newFinding(analysisID, verdict.Kind, verdict.Code, verdict.Message, components, []string{"policy:" + verdict.RuleID}, now))
	}
	for _, from := range ordered {
		source, _ := graph.Component(from)
		for _, to := range graph.Dependencies(from) {
			dependency, _ := graph.Component(to)
			kind, msg, waivable := policy.PairCompatibility(source.License, dependency.License)
			if kind == model.FindingObligation && !policy.IsCopyleft(source.License) && !policy.IsCopyleft(dependency.License) {
				continue
			}
			derivation := []string{"component:" + source.ID, "dependency:" + dependency.ID}
			if waivable {
				derivation = append(derivation, "waiver-permitted")
			}
			findings = append(findings, newFinding(analysisID, kind, "dependency.compatibility", msg, []string{source.ID, dependency.ID}, derivation, now))
		}
	}
	findings = deduplicate(findings)
	return Result{Findings: findings, Ordered: ordered}, nil
}

func componentFindings(component model.Component, analysisID string, now time.Time) []model.Finding {
	findings := []model.Finding{}
	if strings.TrimSpace(component.License) == "" {
		findings = append(findings, newFinding(analysisID, model.FindingWarning, "component.missing-license", "component has no declared license", []string{component.ID}, []string{"component:" + component.ID}, now))
		return findings
	}
	if component.Status == model.ComponentDisputed {
		findings = append(findings, newFinding(analysisID, model.FindingBlocker, "component.disputed", "component license declaration is disputed", []string{component.ID}, []string{"component:" + component.ID}, now))
	}
	if component.Status == model.ComponentRevoked {
		findings = append(findings, newFinding(analysisID, model.FindingBlocker, "component.revoked", "component declaration was revoked", []string{component.ID}, []string{"component:" + component.ID}, now))
	}
	if policy.Lookup(component.License).Family == policy.FamilyUnknown {
		findings = append(findings, newFinding(analysisID, model.FindingWarning, "component.unknown-license", "license is unknown to the compatibility catalog", []string{component.ID}, []string{"component:" + component.ID}, now))
	}
	return findings
}

func componentsForLicenses(graph *Graph, licenses []string) []string {
	out := []string{}
	allowed := map[string]bool{}
	for _, license := range licenses {
		allowed[policy.Normalize(license)] = true
	}
	for _, id := range graph.TopologicalMust() {
		component, _ := graph.Component(id)
		if allowed[policy.Normalize(component.License)] {
			out = append(out, id)
		}
	}
	return out
}
func (g *Graph) TopologicalMust() []string {
	ordered, err := g.Topological()
	if err != nil {
		return nil
	}
	return ordered
}
func newFinding(analysisID string, kind model.FindingKind, code, message string, components, derivation []string, now time.Time) model.Finding {
	return model.Finding{ID: FindingID(analysisID, code, components), AnalysisID: analysisID, Kind: kind, Code: code, Message: message, Components: components, Derivation: derivation, CreatedAt: now.UTC()}
}
func FindingID(analysisID, code string, components []string) string {
	copy := append([]string(nil), components...)
	sort.Strings(copy)
	return analysisID + ":" + code + ":" + strings.Join(copy, ",")
}
func deduplicate(values []model.Finding) []model.Finding {
	seen := map[string]bool{}
	out := make([]model.Finding, 0, len(values))
	for _, v := range values {
		key := fmt.Sprintf("%s:%s", v.Code, strings.Join(v.Components, ","))
		if !seen[key] {
			seen[key] = true
			out = append(out, v)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
