package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/repository"
)

var _ repository.Store = (*Store)(nil)

// Store is a process-local repository. It is the Phase 1 adapter so the API
// can run with no AWS credentials or cloud services.
type Store struct {
	mu          sync.RWMutex
	services    map[domain.ServiceID]domain.Service
	deps        map[string]domain.Dependency
	releases    map[domain.ReleaseID]domain.Release
	deployments map[domain.ReleaseID]domain.Deployment
	commits     map[domain.ReleaseID]domain.Commit
	ciRuns      map[domain.ReleaseID]domain.CIRun
	health      map[domain.HealthSnapshotID]domain.HealthSnapshot
	incidents   map[domain.IncidentID]domain.Incident
	events      map[domain.EventID]domain.ReleaseEvent
	decisions   map[domain.DecisionID]domain.ReleaseDecision
	risks       map[domain.ReleaseID]domain.RiskAssessment
	scans       map[domain.ReleaseID]domain.SecurityScan
	healthCmp   map[domain.ReleaseID]domain.HealthAssessment
	policies    map[domain.ReleaseID]domain.PolicyEvaluation
}

func New() *Store {
	return &Store{
		services:    map[domain.ServiceID]domain.Service{},
		deps:        map[string]domain.Dependency{},
		releases:    map[domain.ReleaseID]domain.Release{},
		deployments: map[domain.ReleaseID]domain.Deployment{},
		commits:     map[domain.ReleaseID]domain.Commit{},
		ciRuns:      map[domain.ReleaseID]domain.CIRun{},
		health:      map[domain.HealthSnapshotID]domain.HealthSnapshot{},
		incidents:   map[domain.IncidentID]domain.Incident{},
		events:      map[domain.EventID]domain.ReleaseEvent{},
		decisions:   map[domain.DecisionID]domain.ReleaseDecision{},
		risks:       map[domain.ReleaseID]domain.RiskAssessment{},
		scans:       map[domain.ReleaseID]domain.SecurityScan{},
		healthCmp:   map[domain.ReleaseID]domain.HealthAssessment{},
		policies:    map[domain.ReleaseID]domain.PolicyEvaluation{},
	}
}

func (s *Store) CreateService(_ context.Context, svc domain.Service) error {
	svc = svc.Normalized()
	if err := svc.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[svc.ID]; ok {
		return domain.AlreadyExistsError{Resource: "service", ID: svc.ID.String()}
	}
	s.services[svc.ID] = svc
	return nil
}

func (s *Store) GetService(_ context.Context, id domain.ServiceID) (domain.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	svc, ok := s.services[id]
	if !ok {
		return domain.Service{}, domain.NotFoundError{Resource: "service", ID: id.String()}
	}
	return svc, nil
}

func (s *Store) ListServices(_ context.Context) ([]domain.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Service, 0, len(s.services))
	for _, svc := range s.services {
		out = append(out, svc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Store) CreateDependency(_ context.Context, d domain.Dependency) error {
	d = d.Normalized()
	if err := d.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.deps[d.Key()]; ok {
		return domain.AlreadyExistsError{Resource: "dependency", ID: d.Key()}
	}
	s.deps[d.Key()] = d
	return nil
}

func (s *Store) ListDependenciesFrom(_ context.Context, from domain.ServiceID) ([]domain.Dependency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Dependency
	for _, d := range s.deps {
		if d.From == from {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].To.String() < out[j].To.String() })
	return out, nil
}

func (s *Store) ListDependenciesTo(_ context.Context, to domain.ServiceID) ([]domain.Dependency, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Dependency
	for _, d := range s.deps {
		if d.To == to {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].From.String() < out[j].From.String() })
	return out, nil
}

func (s *Store) CreateRelease(_ context.Context, r domain.Release) error {
	r = r.Normalized()
	if err := r.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.releases[r.ID]; ok {
		return domain.AlreadyExistsError{Resource: "release", ID: r.ID.String()}
	}
	s.releases[r.ID] = r
	return nil
}

func (s *Store) GetRelease(_ context.Context, id domain.ReleaseID) (domain.Release, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.releases[id]
	if !ok {
		return domain.Release{}, domain.NotFoundError{Resource: "release", ID: id.String()}
	}
	return r, nil
}

func (s *Store) ListReleases(_ context.Context) ([]domain.Release, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Release, 0, len(s.releases))
	for _, r := range s.releases {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *Store) ListReleasesByService(_ context.Context, serviceID domain.ServiceID) ([]domain.Release, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Release
	for _, r := range s.releases {
		if r.ServiceID == serviceID {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *Store) CreateDeployment(_ context.Context, d domain.Deployment) error {
	d = d.Normalized()
	if err := d.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.deployments[d.ReleaseID]; ok {
		return domain.AlreadyExistsError{Resource: "deployment", ID: d.ID.String()}
	}
	s.deployments[d.ReleaseID] = d
	return nil
}

func (s *Store) GetDeploymentByRelease(_ context.Context, releaseID domain.ReleaseID) (domain.Deployment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.deployments[releaseID]
	if !ok {
		return domain.Deployment{}, domain.NotFoundError{Resource: "deployment", ID: releaseID.String()}
	}
	return d, nil
}

func (s *Store) CreateCommit(_ context.Context, c domain.Commit) error {
	c = c.Normalized()
	if err := c.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.commits[c.ReleaseID]; ok {
		return domain.AlreadyExistsError{Resource: "commit", ID: c.SHA.String()}
	}
	s.commits[c.ReleaseID] = c
	return nil
}

func (s *Store) GetCommitByRelease(_ context.Context, releaseID domain.ReleaseID) (domain.Commit, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.commits[releaseID]
	if !ok {
		return domain.Commit{}, domain.NotFoundError{Resource: "commit", ID: releaseID.String()}
	}
	return c, nil
}

func (s *Store) CreateCIRun(_ context.Context, r domain.CIRun) error {
	r = r.Normalized()
	if err := r.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.ciRuns[r.ReleaseID]; ok {
		return domain.AlreadyExistsError{Resource: "ci_run", ID: r.ID.String()}
	}
	s.ciRuns[r.ReleaseID] = r
	return nil
}

func (s *Store) GetCIRunByRelease(_ context.Context, releaseID domain.ReleaseID) (domain.CIRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.ciRuns[releaseID]
	if !ok {
		return domain.CIRun{}, domain.NotFoundError{Resource: "ci_run", ID: releaseID.String()}
	}
	return r, nil
}

func (s *Store) CreateHealthSnapshot(_ context.Context, h domain.HealthSnapshot) error {
	h = h.Normalized()
	if err := h.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.health[h.ID]; ok {
		return domain.AlreadyExistsError{Resource: "health_snapshot", ID: h.ID.String()}
	}
	s.health[h.ID] = h
	return nil
}

func (s *Store) CreateIncident(_ context.Context, i domain.Incident) error {
	i = i.Normalized()
	if err := i.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.incidents[i.ID]; ok {
		return domain.AlreadyExistsError{Resource: "incident", ID: i.ID.String()}
	}
	s.incidents[i.ID] = i
	return nil
}

func (s *Store) CreateEvent(_ context.Context, e domain.ReleaseEvent) error {
	e = e.Normalized()
	if err := e.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.events[e.ID]; ok {
		return domain.AlreadyExistsError{Resource: "event", ID: e.ID.String()}
	}
	s.events[e.ID] = e
	return nil
}

func (s *Store) GetEvent(_ context.Context, id domain.EventID) (domain.ReleaseEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.events[id]
	if !ok {
		return domain.ReleaseEvent{}, domain.NotFoundError{Resource: "event", ID: id.String()}
	}
	return e, nil
}

func (s *Store) ListEventsByRelease(_ context.Context, releaseID domain.ReleaseID) ([]domain.ReleaseEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.ReleaseEvent
	for _, e := range s.events {
		if e.ReleaseID != nil && *e.ReleaseID == releaseID {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID.String() < out[j].ID.String()
		}
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out, nil
}

func (s *Store) CreateDecision(_ context.Context, d domain.ReleaseDecision) error {
	d = d.Normalized()
	if err := d.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.decisions[d.ID]; ok {
		return domain.AlreadyExistsError{Resource: "decision", ID: d.ID.String()}
	}
	s.decisions[d.ID] = d
	return nil
}

func (s *Store) PutRiskAssessment(_ context.Context, a domain.RiskAssessment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.risks[a.ReleaseID] = a
	return nil
}

func (s *Store) GetRiskAssessment(_ context.Context, releaseID domain.ReleaseID) (domain.RiskAssessment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.risks[releaseID]
	if !ok {
		return domain.RiskAssessment{}, domain.NotFoundError{Resource: "risk", ID: releaseID.String()}
	}
	return a, nil
}

func (s *Store) CreateSecurityScan(_ context.Context, scan domain.SecurityScan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.scans[scan.ReleaseID]; ok {
		return domain.AlreadyExistsError{Resource: "security_scan", ID: scan.ID}
	}
	s.scans[scan.ReleaseID] = scan
	return nil
}

func (s *Store) GetSecurityScanByRelease(_ context.Context, releaseID domain.ReleaseID) (domain.SecurityScan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scan, ok := s.scans[releaseID]
	if !ok {
		return domain.SecurityScan{}, domain.NotFoundError{Resource: "security_scan", ID: releaseID.String()}
	}
	return scan, nil
}

func (s *Store) ListIncidentsByService(_ context.Context, serviceID domain.ServiceID) ([]domain.Incident, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Incident
	for _, i := range s.incidents {
		if i.ServiceID == serviceID {
			out = append(out, i)
		}
	}
	return out, nil
}

func (s *Store) ListHealthSnapshotsByService(_ context.Context, serviceID domain.ServiceID) ([]domain.HealthSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.HealthSnapshot
	for _, h := range s.health {
		if h.ServiceID == serviceID {
			out = append(out, h)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WindowStart.Before(out[j].WindowStart) })
	return out, nil
}

func (s *Store) PutHealthComparison(_ context.Context, a domain.HealthAssessment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthCmp[a.ReleaseID] = a
	return nil
}

func (s *Store) GetHealthComparison(_ context.Context, releaseID domain.ReleaseID) (domain.HealthAssessment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.healthCmp[releaseID]
	if !ok {
		return domain.HealthAssessment{}, domain.NotFoundError{Resource: "health", ID: releaseID.String()}
	}
	return a, nil
}

func (s *Store) PutPolicyEvaluation(_ context.Context, e domain.PolicyEvaluation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[e.ReleaseID] = e
	return nil
}

func (s *Store) GetPolicyEvaluation(_ context.Context, releaseID domain.ReleaseID) (domain.PolicyEvaluation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.policies[releaseID]
	if !ok {
		return domain.PolicyEvaluation{}, domain.NotFoundError{Resource: "policy", ID: releaseID.String()}
	}
	return e, nil
}
