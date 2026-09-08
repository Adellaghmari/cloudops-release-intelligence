package service

import (
	"context"
	"strings"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/repository"
)

// Catalog is the application service for services and releases.
// Risk, health, graph, and policy engines must not be implemented here later
// as Gin handlers; they will be sibling packages invoked from services like this.
type Catalog struct {
	store repository.Store
}

func NewCatalog(store repository.Store) *Catalog {
	return &Catalog{store: store}
}

func (c *Catalog) ListServices(ctx context.Context, source *domain.DataSource) ([]domain.Service, error) {
	list, err := c.store.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return list, nil
	}
	out := make([]domain.Service, 0, len(list))
	for _, s := range list {
		if s.Source == *source {
			out = append(out, s)
		}
	}
	return out, nil
}

func (c *Catalog) GetService(ctx context.Context, id domain.ServiceID) (ServiceDetail, error) {
	svc, err := c.store.GetService(ctx, id)
	if err != nil {
		return ServiceDetail{}, err
	}
	from, err := c.store.ListDependenciesFrom(ctx, id)
	if err != nil {
		return ServiceDetail{}, err
	}
	to, err := c.store.ListDependenciesTo(ctx, id)
	if err != nil {
		return ServiceDetail{}, err
	}
	return ServiceDetail{Service: svc, DependsOn: from, DependedBy: to}, nil
}

func (c *Catalog) ListReleases(ctx context.Context, source *domain.DataSource, serviceID *domain.ServiceID) ([]domain.Release, error) {
	var (
		list []domain.Release
		err  error
	)
	if serviceID != nil {
		list, err = c.store.ListReleasesByService(ctx, *serviceID)
	} else {
		list, err = c.store.ListReleases(ctx)
	}
	if err != nil {
		return nil, err
	}
	if source == nil {
		return list, nil
	}
	out := make([]domain.Release, 0, len(list))
	for _, r := range list {
		if r.Source == *source {
			out = append(out, r)
		}
	}
	return out, nil
}

func (c *Catalog) GetRelease(ctx context.Context, id domain.ReleaseID) (ReleaseDetail, error) {
	rel, err := c.store.GetRelease(ctx, id)
	if err != nil {
		return ReleaseDetail{}, err
	}
	svc, err := c.store.GetService(ctx, rel.ServiceID)
	if err != nil {
		return ReleaseDetail{}, err
	}
	commit, err := optional(c.store.GetCommitByRelease(ctx, id))
	if err != nil {
		return ReleaseDetail{}, err
	}
	deploy, err := optional(c.store.GetDeploymentByRelease(ctx, id))
	if err != nil {
		return ReleaseDetail{}, err
	}
	ci, err := optional(c.store.GetCIRunByRelease(ctx, id))
	if err != nil {
		return ReleaseDetail{}, err
	}
	return ReleaseDetail{
		Release:    rel,
		Service:    svc,
		Commit:     commit,
		Deployment: deploy,
		CIRun:      ci,
	}, nil
}

type ServiceDetail struct {
	Service    domain.Service
	DependsOn  []domain.Dependency
	DependedBy []domain.Dependency
}

type ReleaseDetail struct {
	Release    domain.Release
	Service    domain.Service
	Commit     *domain.Commit
	Deployment *domain.Deployment
	CIRun      *domain.CIRun
}

func ParseSourceQuery(raw string) (*domain.DataSource, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	src, err := domain.ParseDataSource(raw)
	if err != nil {
		return nil, err
	}
	return &src, nil
}

func optional[T any](v T, err error) (*T, error) {
	if err == nil {
		return &v, nil
	}
	if domain.IsNotFound(err) {
		return nil, nil
	}
	return nil, err
}
