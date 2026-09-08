package service

import (
	"context"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/graph"
)

func (c *Catalog) Impact(ctx context.Context, id domain.ReleaseID) (graph.Result, error) {
	detail, err := c.GetRelease(ctx, id)
	if err != nil {
		return graph.Result{}, err
	}
	return c.ImpactForService(ctx, detail.Service.ID)
}

func (c *Catalog) ImpactForService(ctx context.Context, id domain.ServiceID) (graph.Result, error) {
	services, err := c.store.ListServices(ctx)
	if err != nil {
		return graph.Result{}, err
	}
	var edges []graph.Edge
	for _, svc := range services {
		deps, err := c.store.ListDependenciesFrom(ctx, svc.ID)
		if err != nil {
			return graph.Result{}, err
		}
		for _, d := range deps {
			edges = append(edges, graph.Edge{From: d.From, To: d.To})
		}
	}
	return graph.BlastRadius(id, services, edges), nil
}
