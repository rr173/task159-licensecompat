package analysis

import (
	"sort"

	"github.com/rr173/task159-licensecompat/internal/model"
)

type Graph struct {
	nodes    map[string]model.Component
	outgoing map[string][]string
	incoming map[string][]string
}

func BuildGraph(submission model.Submission) (*Graph, error) {
	g := &Graph{nodes: map[string]model.Component{}, outgoing: map[string][]string{}, incoming: map[string][]string{}}
	for _, component := range submission.Components {
		g.nodes[component.ID] = component
		g.outgoing[component.ID] = nil
		g.incoming[component.ID] = nil
	}
	for _, edge := range submission.Edges {
		g.outgoing[edge.From] = append(g.outgoing[edge.From], edge.To)
		g.incoming[edge.To] = append(g.incoming[edge.To], edge.From)
	}
	if _, err := g.Topological(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *Graph) Topological() ([]string, error) {
	degree := map[string]int{}
	ready := make([]string, 0)
	for id := range g.nodes {
		degree[id] = len(g.incoming[id])
		if degree[id] == 0 {
			ready = append(ready, id)
		}
	}
	sort.Strings(ready)
	ordered := make([]string, 0, len(g.nodes))
	for len(ready) > 0 {
		id := ready[0]
		ready = ready[1:]
		ordered = append(ordered, id)
		for _, to := range g.outgoing[id] {
			degree[to]--
			if degree[to] == 0 {
				ready = append(ready, to)
				sort.Strings(ready)
			}
		}
	}
	if len(ordered) != len(g.nodes) {
		return nil, model.ErrCircularGraph
	}
	return ordered, nil
}

func (g *Graph) Component(id string) (model.Component, bool) { c, ok := g.nodes[id]; return c, ok }
func (g *Graph) Dependencies(id string) []string {
	out := append([]string(nil), g.outgoing[id]...)
	sort.Strings(out)
	return out
}
func (g *Graph) Dependents(id string) []string {
	out := append([]string(nil), g.incoming[id]...)
	sort.Strings(out)
	return out
}
func (g *Graph) Roots() []string {
	out := []string{}
	for id := range g.nodes {
		if len(g.incoming[id]) == 0 {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}
