package service

import (
	"context"
	"sort"

	"licensecompat.local/internal/model"
)

type ComponentReport struct {
	ComponentID string                `json:"component_id"`
	Name        string                `json:"name"`
	License     string                `json:"license"`
	Status      model.ComponentStatus `json:"status"`
	Findings    []model.Finding       `json:"findings"`
}

type DecisionReport struct {
	Decision   model.Decision    `json:"decision"`
	Summary    model.Summary     `json:"summary"`
	Components []ComponentReport `json:"components"`
	Published  bool              `json:"published"`
}

func (s *Service) Report(ctx context.Context, analysisID string) (DecisionReport, error) {
	decision, err := s.Decision(ctx, analysisID)
	if err != nil {
		return DecisionReport{}, err
	}
	submission, err := s.store.Submission(ctx, decision.Analysis.SubmissionID)
	if err != nil {
		return DecisionReport{}, err
	}
	byComponent := make(map[string][]model.Finding, len(submission.Components))
	for _, finding := range decision.Findings {
		for _, componentID := range finding.Components {
			byComponent[componentID] = append(byComponent[componentID], finding)
		}
	}
	components := make([]ComponentReport, 0, len(submission.Components))
	for _, component := range submission.Components {
		findings := append([]model.Finding(nil), byComponent[component.ID]...)
		sort.Slice(findings, func(i, j int) bool {
			if findings[i].Kind != findings[j].Kind {
				return findings[i].Kind < findings[j].Kind
			}
			return findings[i].Code < findings[j].Code
		})
		components = append(components, ComponentReport{ComponentID: component.ID, Name: component.Name, License: component.License, Status: component.Status, Findings: findings})
	}
	sort.Slice(components, func(i, j int) bool { return components[i].ComponentID < components[j].ComponentID })
	return DecisionReport{Decision: decision, Summary: model.Summarize(decision.Findings, submission.Components), Components: components, Published: decision.Analysis.Status == model.AnalysisPublished}, nil
}
