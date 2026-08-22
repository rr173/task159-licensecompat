package analysis

import "sort"

func (g *Graph) Path(from, to string) ([]string, bool) {
	if _, ok := g.Component(from); !ok {
		return nil, false
	}
	if _, ok := g.Component(to); !ok {
		return nil, false
	}
	queue := [][]string{{from}}
	visited := map[string]bool{from: true}
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		last := path[len(path)-1]
		if last == to {
			return path, true
		}
		next := g.Dependencies(last)
		sort.Strings(next)
		for _, child := range next {
			if !visited[child] {
				visited[child] = true
				clone := append([]string(nil), path...)
				queue = append(queue, append(clone, child))
			}
		}
	}
	return nil, false
}

func (g *Graph) Closure(start string) []string {
	seen := map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		for _, child := range g.Dependencies(id) {
			if !seen[child] {
				seen[child] = true
				visit(child)
			}
		}
	}
	visit(start)
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
