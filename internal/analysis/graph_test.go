package analysis

import (
	"errors"
	"licensecompat.local/internal/model"
	"testing"
)

func TestGraphPathAndClosure(t *testing.T) {
	g, err := BuildGraph(model.Submission{Components: []model.Component{{ID: "app"}, {ID: "mid"}, {ID: "lib"}}, Edges: []model.Edge{{From: "app", To: "mid"}, {From: "mid", To: "lib"}}})
	if err != nil {
		t.Fatal(err)
	}
	path, ok := g.Path("app", "lib")
	if !ok || len(path) != 3 {
		t.Fatalf("path=%v ok=%v", path, ok)
	}
	if got := g.Closure("app"); len(got) != 2 {
		t.Fatalf("closure=%v", got)
	}
}
func TestGraphRejectsCycle(t *testing.T) {
	_, err := BuildGraph(model.Submission{Components: []model.Component{{ID: "a"}, {ID: "b"}}, Edges: []model.Edge{{From: "a", To: "b"}, {From: "b", To: "a"}}})
	if !errors.Is(err, model.ErrCircularGraph) {
		t.Fatalf("expected cycle, got %v", err)
	}
}
