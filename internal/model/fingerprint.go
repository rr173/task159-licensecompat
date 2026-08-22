package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

func CanonicalLicense(value string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))
	switch normalized {
	case "GPL3", "GPLV3", "GPL3.0":
		return "GPL-3.0"
	case "AGPL3", "AGPLV3", "AGPL3.0":
		return "AGPL-3.0"
	case "APACHE2.0", "APACHE-2":
		return "APACHE-2.0"
	case "MITLICENSE":
		return "MIT"
	case "BSD3":
		return "BSD-3-CLAUSE"
	default:
		return normalized
	}
}

func SubmissionFingerprint(s Submission) (string, error) {
	copy := s
	copy.ID = ""
	copy.Fingerprint = ""
	copy.CreatedAt = time.Time{}
	sort.Slice(copy.Components, func(i, j int) bool { return copy.Components[i].ID < copy.Components[j].ID })
	sort.Slice(copy.Edges, func(i, j int) bool {
		if copy.Edges[i].From != copy.Edges[j].From {
			return copy.Edges[i].From < copy.Edges[j].From
		}
		return copy.Edges[i].To < copy.Edges[j].To
	})
	for i := range copy.Components {
		copy.Components[i].License = strings.TrimSpace(copy.Components[i].License)
		if copy.Components[i].Metadata == nil {
			copy.Components[i].Metadata = map[string]string{}
		}
	}
	b, err := json.Marshal(copy)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func Snapshot(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Restore[T any](snapshot string) (T, error) {
	var value T
	err := json.Unmarshal([]byte(snapshot), &value)
	return value, err
}
