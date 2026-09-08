package repository

import (
	"context"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

// Store is the persistence port. Implementations must not leak AWS or HTTP types.
type Store interface {
	CreateService(ctx context.Context, s domain.Service) error
	GetService(ctx context.Context, id domain.ServiceID) (domain.Service, error)
	ListServices(ctx context.Context) ([]domain.Service, error)

	CreateDependency(ctx context.Context, d domain.Dependency) error
	ListDependenciesFrom(ctx context.Context, from domain.ServiceID) ([]domain.Dependency, error)
	ListDependenciesTo(ctx context.Context, to domain.ServiceID) ([]domain.Dependency, error)

	CreateRelease(ctx context.Context, r domain.Release) error
	GetRelease(ctx context.Context, id domain.ReleaseID) (domain.Release, error)
	ListReleases(ctx context.Context) ([]domain.Release, error)
	ListReleasesByService(ctx context.Context, serviceID domain.ServiceID) ([]domain.Release, error)

	CreateDeployment(ctx context.Context, d domain.Deployment) error
	GetDeploymentByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.Deployment, error)

	CreateCommit(ctx context.Context, c domain.Commit) error
	GetCommitByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.Commit, error)

	CreateCIRun(ctx context.Context, r domain.CIRun) error
	GetCIRunByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.CIRun, error)

	CreateHealthSnapshot(ctx context.Context, h domain.HealthSnapshot) error
	CreateIncident(ctx context.Context, i domain.Incident) error
	CreateEvent(ctx context.Context, e domain.ReleaseEvent) error
	CreateDecision(ctx context.Context, d domain.ReleaseDecision) error
	GetEvent(ctx context.Context, id domain.EventID) (domain.ReleaseEvent, error)
	ListEventsByRelease(ctx context.Context, releaseID domain.ReleaseID) ([]domain.ReleaseEvent, error)
	PutRiskAssessment(ctx context.Context, a domain.RiskAssessment) error
	GetRiskAssessment(ctx context.Context, releaseID domain.ReleaseID) (domain.RiskAssessment, error)
	CreateSecurityScan(ctx context.Context, s domain.SecurityScan) error
	GetSecurityScanByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.SecurityScan, error)
	ListIncidentsByService(ctx context.Context, serviceID domain.ServiceID) ([]domain.Incident, error)
	ListHealthSnapshotsByService(ctx context.Context, serviceID domain.ServiceID) ([]domain.HealthSnapshot, error)
	PutHealthComparison(ctx context.Context, a domain.HealthAssessment) error
	GetHealthComparison(ctx context.Context, releaseID domain.ReleaseID) (domain.HealthAssessment, error)
}
