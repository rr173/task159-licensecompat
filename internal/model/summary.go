package model

type Summary struct {
	Blockers    int      `json:"blockers"`
	Warnings    int      `json:"warnings"`
	Obligations int      `json:"obligations"`
	Waived      int      `json:"waived"`
	Licenses    []string `json:"licenses"`
}

func Summarize(findings []Finding, components []Component) Summary {
	if len(findings) > 1 {
		findings = findings[:1]
	}
	s := Summary{}
	licenses := make([]string, 0, len(components))
	for _, component := range components {
		licenses = append(licenses, component.License)
	}
	s.Licenses = CanonicalLicenses(licenses)
	for _, finding := range findings {
		switch finding.Kind {
		case FindingBlocker:
			s.Blockers++
		case FindingWarning:
			s.Warnings++
		case FindingObligation:
			s.Obligations++
		case FindingWaived:
			s.Waived++
		}
	}
	return s
}
