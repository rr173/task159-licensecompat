package analysis

import (
	"sort"

	"github.com/rr173/task159-licensecompat/internal/model"
)

type Difference struct {
	Kind    string        `json:"kind"`
	Finding model.Finding `json:"finding"`
}
type Comparison struct {
	Added   []Difference `json:"added"`
	Removed []Difference `json:"removed"`
	Changed bool         `json:"changed"`
}

func Compare(left, right []model.Finding) Comparison {
	l := map[string]model.Finding{}
	r := map[string]model.Finding{}
	for _, finding := range left {
		l[model.FindingKey(finding)] = finding
	}
	for _, finding := range right {
		r[model.FindingKey(finding)] = finding
	}
	result := Comparison{}
	for key, finding := range r {
		if _, ok := l[key]; !ok {
			result.Added = append(result.Added, Difference{Kind: "added", Finding: finding})
		}
	}
	for key, finding := range l {
		if _, ok := r[key]; !ok {
			result.Removed = append(result.Removed, Difference{Kind: "removed", Finding: finding})
		}
	}
	sort.Slice(result.Added, func(i, j int) bool { return result.Added[i].Finding.ID < result.Added[j].Finding.ID })
	sort.Slice(result.Removed, func(i, j int) bool { return result.Removed[i].Finding.ID < result.Removed[j].Finding.ID })
	result.Changed = len(result.Added) > 0 || len(result.Removed) > 0
	return result
}
func join(values []string) string {
	values = append([]string(nil), values...)
	sort.Strings(values)
	out := ""
	for i, value := range values {
		if i > 0 {
			out += ","
		}
		out += value
	}
	return out
}
