package graph

import (
	"testing"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

func svc(id string) domain.Service { return domain.Service{ID: domain.ServiceID(id)} }

func TestChainAndUnknown(t *testing.T) {
	svcs := []domain.Service{svc("a"), svc("b"), svc("c")}
	edges := []Edge{{"a", "b"}, {"b", "c"}}
	r := BlastRadius("c", svcs, edges)
	if len(r.DirectDependents) != 1 || r.DirectDependents[0] != "b" {
		t.Fatalf("direct=%v", r.DirectDependents)
	}
	if len(r.TransitiveDependents) != 1 || r.TransitiveDependents[0] != "a" {
		t.Fatalf("trans=%v", r.TransitiveDependents)
	}
	if BlastRadius("z", svcs, edges).Unknown != true {
		t.Fatal("unknown")
	}
}

func TestBranchingAndDepths(t *testing.T) {
	svcs := []domain.Service{svc("root"), svc("left"), svc("right"), svc("leaf")}
	edges := []Edge{{"left", "root"}, {"right", "root"}, {"leaf", "left"}}
	r := BlastRadius("root", svcs, edges)
	if r.MaxDepth < 2 {
		t.Fatalf("depth=%d", r.MaxDepth)
	}
	if len(r.DirectDependents) != 2 {
		t.Fatalf("direct=%v", r.DirectDependents)
	}
	found := false
	for _, id := range r.TransitiveDependents {
		if id == "leaf" {
			found = true
		}
	}
	if !found {
		t.Fatalf("trans=%v", r.TransitiveDependents)
	}
}

func TestDiamondCycleIsolated(t *testing.T) {
	svcs := []domain.Service{svc("s"), svc("a"), svc("b"), svc("c"), svc("iso")}
	edges := []Edge{{"a", "s"}, {"b", "s"}, {"c", "a"}, {"c", "b"}, {"a", "b"}, {"b", "a"}}
	r := BlastRadius("s", svcs, edges)
	if len(r.DirectDependents) < 2 {
		t.Fatalf("diamond direct=%v", r.DirectDependents)
	}
	if len(r.Cycles) == 0 {
		t.Fatal("expected cycle")
	}
	iso := BlastRadius("iso", svcs, edges)
	if len(iso.DirectDependents)+len(iso.TransitiveDependents) != 0 {
		t.Fatalf("isolated %+v", iso)
	}
}
