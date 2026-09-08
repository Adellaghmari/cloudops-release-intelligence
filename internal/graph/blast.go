package graph

import "github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"

type Edge struct {
	From domain.ServiceID
	To   domain.ServiceID
}

type Node struct {
	ID    domain.ServiceID
	Role  string
	Depth int
}

type Result struct {
	ChangedService       domain.ServiceID
	DirectDependents     []domain.ServiceID
	TransitiveDependents []domain.ServiceID
	Upstream             []domain.ServiceID
	CriticalInRadius     []domain.ServiceID
	Nodes                []Node
	Edges                []Edge
	MaxDepth             int
	Cycles               [][]domain.ServiceID
	Unknown              bool
	Empty                bool
	Algorithm            string
	Disclaimer           string
}

func BlastRadius(changed domain.ServiceID, services []domain.Service, edges []Edge) Result {
	known := map[domain.ServiceID]struct{}{}
	for _, s := range services {
		known[s.ID] = struct{}{}
	}
	disclaimer := "Potential blast radius only. Reachability is not observed breakage."
	if _, ok := known[changed]; !ok {
		return Result{ChangedService: changed, Unknown: true, Algorithm: "bfs", Disclaimer: disclaimer}
	}
	if len(edges) == 0 {
		return Result{ChangedService: changed, Empty: true, Algorithm: "bfs", Disclaimer: disclaimer}
	}
	dependsOn := map[domain.ServiceID][]domain.ServiceID{}
	dependedBy := map[domain.ServiceID][]domain.ServiceID{}
	for _, e := range edges {
		dependsOn[e.From] = append(dependsOn[e.From], e.To)
		dependedBy[e.To] = append(dependedBy[e.To], e.From)
	}
	direct, trans, depths, depth := bfs(changed, dependedBy)
	up, _, upDepths, _ := bfs(changed, dependsOn)
	crit := criticalIn(append(append([]domain.ServiceID{}, direct...), trans...), services)
	nodes := []Node{{ID: changed, Role: "changed", Depth: 0}}
	for _, id := range direct {
		nodes = append(nodes, Node{ID: id, Role: "direct_dependent", Depth: depths[id]})
	}
	for _, id := range trans {
		nodes = append(nodes, Node{ID: id, Role: "transitive_dependent", Depth: depths[id]})
	}
	for _, id := range up {
		nodes = append(nodes, Node{ID: id, Role: "upstream", Depth: -upDepths[id]})
	}
	return Result{
		ChangedService: changed, DirectDependents: direct, TransitiveDependents: trans,
		Upstream: up, CriticalInRadius: crit, Nodes: nodes, Edges: edges,
		MaxDepth: depth, Cycles: cycles(dependsOn), Algorithm: "bfs", Disclaimer: disclaimer,
	}
}

func criticalIn(ids []domain.ServiceID, services []domain.Service) []domain.ServiceID {
	crit := map[domain.ServiceID]struct{}{}
	for _, s := range services {
		if s.Criticality == domain.CriticalityHigh || s.Criticality == domain.CriticalityCritical {
			crit[s.ID] = struct{}{}
		}
	}
	var out []domain.ServiceID
	seen := map[domain.ServiceID]struct{}{}
	for _, id := range ids {
		if _, ok := crit[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func bfs(start domain.ServiceID, adj map[domain.ServiceID][]domain.ServiceID) (direct, trans []domain.ServiceID, depths map[domain.ServiceID]int, maxDepth int) {
	depths = map[domain.ServiceID]int{}
	type node struct {
		id    domain.ServiceID
		depth int
	}
	seen := map[domain.ServiceID]struct{}{start: {}}
	q := []node{{start, 0}}
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		for _, n := range adj[cur.id] {
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			d := cur.depth + 1
			if d > maxDepth {
				maxDepth = d
			}
			depths[n] = d
			if d == 1 {
				direct = append(direct, n)
			} else {
				trans = append(trans, n)
			}
			q = append(q, node{n, d})
		}
	}
	return
}

func cycles(adj map[domain.ServiceID][]domain.ServiceID) [][]domain.ServiceID {
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[domain.ServiceID]int{}
	var stack []domain.ServiceID
	var out [][]domain.ServiceID
	var visit func(domain.ServiceID)
	visit = func(n domain.ServiceID) {
		color[n] = gray
		stack = append(stack, n)
		for _, m := range adj[n] {
			switch color[m] {
			case white:
				visit(m)
			case gray:
				cyc := []domain.ServiceID{}
				for i := len(stack) - 1; i >= 0; i-- {
					cyc = append([]domain.ServiceID{stack[i]}, cyc...)
					if stack[i] == m {
						break
					}
				}
				out = append(out, cyc)
			}
		}
		stack = stack[:len(stack)-1]
		color[n] = black
	}
	for n := range adj {
		if color[n] == white {
			visit(n)
		}
	}
	return out
}
