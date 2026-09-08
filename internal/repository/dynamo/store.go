package dynamo

import (
	"context"
	"sort"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/repository"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var _ repository.Store = (*Store)(nil)

type API interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}

type Store struct {
	api   API
	table string
}

func New(api API, table string) *Store {
	return &Store{api: api, table: table}
}

func (s *Store) CreateService(ctx context.Context, svc domain.Service) error {
	svc = svc.Normalized()
	if err := svc.Validate(); err != nil {
		return err
	}
	gsi1pk, gsi1sk := serviceNameGSI1(svc.Name)
	return s.putNew(ctx, "service", record{
		PK: servicePK(svc.ID), SK: metaSK(), GSI1PK: gsi1pk, GSI1SK: gsi1sk,
		EntityType: "SERVICE", Payload: mustPayload(servicePayload{
			ID: svc.ID.String(), Name: svc.Name, Description: svc.Description,
			Criticality: string(svc.Criticality), Source: string(svc.Source),
			CreatedAt: svc.CreatedAt, UpdatedAt: svc.UpdatedAt,
		}),
	}, svc.ID.String())
}

func (s *Store) GetService(ctx context.Context, id domain.ServiceID) (domain.Service, error) {
	rec, err := s.get(ctx, servicePK(id), metaSK())
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.Service{}, domain.NotFoundError{Resource: "service", ID: id.String()}
		}
		return domain.Service{}, err
	}
	var p servicePayload
	if err := decodePayload(rec.Payload, &p); err != nil {
		return domain.Service{}, wrapErr("get_service", err)
	}
	return serviceFrom(p).Normalized(), nil
}

func (s *Store) ListServices(ctx context.Context) ([]domain.Service, error) {
	pk, prefix := typeServiceGSI1()
	items, err := s.queryIndex(ctx, "GSI1", "GSI1PK = :pk AND begins_with(GSI1SK, :sk)", map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: pk},
		":sk": &types.AttributeValueMemberS{Value: prefix},
	}, true)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Service, 0, len(items))
	for _, rec := range items {
		var p servicePayload
		if err := decodePayload(rec.Payload, &p); err != nil {
			return nil, wrapErr("list_services", err)
		}
		out = append(out, serviceFrom(p).Normalized())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *Store) CreateDependency(ctx context.Context, d domain.Dependency) error {
	d = d.Normalized()
	if err := d.Validate(); err != nil {
		return err
	}
	gsi1pk, gsi1sk := depByGSI1(d.To, d.From)
	return s.putNew(ctx, "dependency", record{
		PK: servicePK(d.From), SK: depSK(d.To), GSI1PK: gsi1pk, GSI1SK: gsi1sk,
		EntityType: "DEPENDENCY", Payload: mustPayload(dependencyPayload{
			From: d.From.String(), To: d.To.String(), Kind: string(d.Kind), CreatedAt: d.CreatedAt,
		}),
	}, d.Key())
}

func (s *Store) ListDependenciesFrom(ctx context.Context, from domain.ServiceID) ([]domain.Dependency, error) {
	return s.listDeps(ctx, servicePK(from), "DEP#", false)
}

func (s *Store) ListDependenciesTo(ctx context.Context, to domain.ServiceID) ([]domain.Dependency, error) {
	items, err := s.queryIndex(ctx, "GSI1", "GSI1PK = :pk AND begins_with(GSI1SK, :sk)", map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: servicePK(to)},
		":sk": &types.AttributeValueMemberS{Value: "DEPBY#"},
	}, true)
	if err != nil {
		return nil, err
	}
	return decodeDeps(items, true)
}

func (s *Store) listDeps(ctx context.Context, pk, skPrefix string, byFrom bool) ([]domain.Dependency, error) {
	items, err := s.queryPK(ctx, pk, skPrefix)
	if err != nil {
		return nil, err
	}
	return decodeDeps(items, byFrom)
}

func decodeDeps(items []record, _ bool) ([]domain.Dependency, error) {
	out := make([]domain.Dependency, 0, len(items))
	for _, rec := range items {
		var p dependencyPayload
		if err := decodePayload(rec.Payload, &p); err != nil {
			return nil, wrapErr("list_deps", err)
		}
		out = append(out, domain.Dependency{
			From: domain.ServiceID(p.From), To: domain.ServiceID(p.To),
			Kind: domain.DependencyKind(p.Kind), CreatedAt: p.CreatedAt,
		}.Normalized())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].From == out[j].From {
			return out[i].To.String() < out[j].To.String()
		}
		return out[i].From.String() < out[j].From.String()
	})
	return out, nil
}

func (s *Store) CreateRelease(ctx context.Context, r domain.Release) error {
	r = r.Normalized()
	if err := r.Validate(); err != nil {
		return err
	}
	g1pk, g1sk := releaseByServiceGSI1(r.ServiceID, r.CreatedAt, r.ID)
	g2pk, g2sk := releaseTypeGSI2(r.CreatedAt, r.ID)
	return s.putNew(ctx, "release", record{
		PK: releasePK(r.ID), SK: metaSK(), GSI1PK: g1pk, GSI1SK: g1sk, GSI2PK: g2pk, GSI2SK: g2sk,
		EntityType: "RELEASE", Payload: mustPayload(releasePayload{
			ID: r.ID.String(), ServiceID: r.ServiceID.String(), Version: r.Version,
			GitSHA: r.GitSHA.String(), Environment: string(r.Environment), Status: string(r.Status),
			Source: string(r.Source), Scenario: r.Scenario, CreatedAt: r.CreatedAt,
		}),
	}, r.ID.String())
}

func (s *Store) GetRelease(ctx context.Context, id domain.ReleaseID) (domain.Release, error) {
	rec, err := s.get(ctx, releasePK(id), metaSK())
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.Release{}, domain.NotFoundError{Resource: "release", ID: id.String()}
		}
		return domain.Release{}, err
	}
	var p releasePayload
	if err := decodePayload(rec.Payload, &p); err != nil {
		return domain.Release{}, wrapErr("get_release", err)
	}
	return releaseFrom(p).Normalized(), nil
}

func (s *Store) ListReleases(ctx context.Context) ([]domain.Release, error) {
	items, err := s.queryIndex(ctx, "GSI2", "GSI2PK = :pk", map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: "TYPE#RELEASE"},
	}, false)
	if err != nil {
		return nil, err
	}
	return decodeReleases(items)
}

func (s *Store) ListReleasesByService(ctx context.Context, serviceID domain.ServiceID) ([]domain.Release, error) {
	items, err := s.queryIndex(ctx, "GSI1", "GSI1PK = :pk AND begins_with(GSI1SK, :sk)", map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: servicePK(serviceID)},
		":sk": &types.AttributeValueMemberS{Value: "RELEASE#"},
	}, false)
	if err != nil {
		return nil, err
	}
	return decodeReleases(items)
}

func decodeReleases(items []record) ([]domain.Release, error) {
	out := make([]domain.Release, 0, len(items))
	for _, rec := range items {
		var p releasePayload
		if err := decodePayload(rec.Payload, &p); err != nil {
			return nil, wrapErr("list_releases", err)
		}
		out = append(out, releaseFrom(p).Normalized())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *Store) CreateDeployment(ctx context.Context, d domain.Deployment) error {
	d = d.Normalized()
	if err := d.Validate(); err != nil {
		return err
	}
	return s.putNew(ctx, "deployment", record{
		PK: releasePK(d.ReleaseID), SK: deploySK(d.ID), EntityType: "DEPLOYMENT",
		Payload: mustPayload(deploymentPayload{
			ID: d.ID.String(), ReleaseID: d.ReleaseID.String(), ServiceID: d.ServiceID.String(),
			Environment: string(d.Environment), Status: string(d.Status), Target: d.Target,
			ImageDigest: d.ImageDigest, ArtifactURI: d.ArtifactURI, StartedAt: d.StartedAt, CompletedAt: d.CompletedAt,
		}),
	}, d.ID.String())
}

func (s *Store) GetDeploymentByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.Deployment, error) {
	items, err := s.queryPK(ctx, releasePK(releaseID), "DEPLOY#")
	if err != nil {
		return domain.Deployment{}, err
	}
	if len(items) == 0 {
		return domain.Deployment{}, domain.NotFoundError{Resource: "deployment", ID: releaseID.String()}
	}
	var p deploymentPayload
	if err := decodePayload(items[0].Payload, &p); err != nil {
		return domain.Deployment{}, wrapErr("get_deployment", err)
	}
	return domain.Deployment{
		ID: domain.DeploymentID(p.ID), ReleaseID: domain.ReleaseID(p.ReleaseID), ServiceID: domain.ServiceID(p.ServiceID),
		Environment: domain.Environment(p.Environment), Status: domain.DeploymentStatus(p.Status),
		Target: p.Target, ImageDigest: p.ImageDigest, ArtifactURI: p.ArtifactURI,
		StartedAt: p.StartedAt, CompletedAt: p.CompletedAt,
	}.Normalized(), nil
}

func (s *Store) CreateCommit(ctx context.Context, c domain.Commit) error {
	c = c.Normalized()
	if err := c.Validate(); err != nil {
		return err
	}
	return s.putNew(ctx, "commit", record{
		PK: releasePK(c.ReleaseID), SK: commitSK(), EntityType: "COMMIT",
		Payload: mustPayload(commitPayload{
			SHA: c.SHA.String(), ReleaseID: c.ReleaseID.String(), ServiceID: c.ServiceID.String(),
			Message: c.Message, Author: c.Author, FilesChanged: c.FilesChanged,
			LinesAdded: c.LinesAdded, LinesDeleted: c.LinesDeleted,
			MigrationPresent: c.MigrationPresent, MigrationReversible: c.MigrationReversible,
			ConfigChangePresent: c.ConfigChangePresent, CommittedAt: c.CommittedAt,
		}),
	}, c.SHA.String())
}

func (s *Store) GetCommitByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.Commit, error) {
	rec, err := s.get(ctx, releasePK(releaseID), commitSK())
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.Commit{}, domain.NotFoundError{Resource: "commit", ID: releaseID.String()}
		}
		return domain.Commit{}, err
	}
	var p commitPayload
	if err := decodePayload(rec.Payload, &p); err != nil {
		return domain.Commit{}, wrapErr("get_commit", err)
	}
	return domain.Commit{
		SHA: domain.CommitSHA(p.SHA), ReleaseID: domain.ReleaseID(p.ReleaseID), ServiceID: domain.ServiceID(p.ServiceID),
		Message: p.Message, Author: p.Author, FilesChanged: p.FilesChanged,
		LinesAdded: p.LinesAdded, LinesDeleted: p.LinesDeleted,
		MigrationPresent: p.MigrationPresent, MigrationReversible: p.MigrationReversible,
		ConfigChangePresent: p.ConfigChangePresent, CommittedAt: p.CommittedAt,
	}.Normalized(), nil
}

func (s *Store) CreateCIRun(ctx context.Context, r domain.CIRun) error {
	r = r.Normalized()
	if err := r.Validate(); err != nil {
		return err
	}
	return s.putNew(ctx, "ci_run", record{
		PK: releasePK(r.ReleaseID), SK: ciSK(r.ID), EntityType: "CI_RUN",
		Payload: mustPayload(ciPayload{
			ID: r.ID.String(), ReleaseID: r.ReleaseID.String(), WorkflowName: r.WorkflowName,
			Status: string(r.Status), FailedTests: r.FailedTests, FailedAttempts: r.FailedAttempts,
			ExternalRunID: r.ExternalRunID, StartedAt: r.StartedAt, CompletedAt: r.CompletedAt,
		}),
	}, r.ID.String())
}

func (s *Store) GetCIRunByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.CIRun, error) {
	items, err := s.queryPK(ctx, releasePK(releaseID), "CI#")
	if err != nil {
		return domain.CIRun{}, err
	}
	if len(items) == 0 {
		return domain.CIRun{}, domain.NotFoundError{Resource: "ci_run", ID: releaseID.String()}
	}
	var p ciPayload
	if err := decodePayload(items[0].Payload, &p); err != nil {
		return domain.CIRun{}, wrapErr("get_ci", err)
	}
	return domain.CIRun{
		ID: domain.CIRunID(p.ID), ReleaseID: domain.ReleaseID(p.ReleaseID), WorkflowName: p.WorkflowName,
		Status: domain.CIRunStatus(p.Status), FailedTests: p.FailedTests, FailedAttempts: p.FailedAttempts,
		ExternalRunID: p.ExternalRunID, StartedAt: p.StartedAt, CompletedAt: p.CompletedAt,
	}.Normalized(), nil
}

func (s *Store) CreateHealthSnapshot(ctx context.Context, h domain.HealthSnapshot) error {
	h = h.Normalized()
	if err := h.Validate(); err != nil {
		return err
	}
	var rid *string
	g1pk, g1sk := "", ""
	if h.ReleaseID != nil {
		v := h.ReleaseID.String()
		rid = &v
		g1pk, g1sk = healthByReleaseGSI1(*h.ReleaseID, h.CapturedAt)
	}
	return s.putNew(ctx, "health_snapshot", record{
		PK: servicePK(h.ServiceID), SK: healthSK(h.WindowEnd, h.ID),
		GSI1PK: g1pk, GSI1SK: g1sk, EntityType: "HEALTH",
		Payload: mustPayload(healthPayload{
			ID: h.ID.String(), ServiceID: h.ServiceID.String(), ReleaseID: rid, WindowKind: string(h.WindowKind),
			WindowStart: h.WindowStart, WindowEnd: h.WindowEnd, RequestCount: h.RequestCount,
			ErrorRate: h.ErrorRate, Availability: h.Availability, LatencyP50MS: h.LatencyP50MS,
			LatencyP95MS: h.LatencyP95MS, LatencyP99MS: h.LatencyP99MS, CPUPct: h.CPUPct,
			MemoryPct: h.MemoryPct, QueueBacklog: h.QueueBacklog, CapturedAt: h.CapturedAt, Source: string(h.Source),
		}),
	}, h.ID.String())
}

func (s *Store) CreateIncident(ctx context.Context, i domain.Incident) error {
	i = i.Normalized()
	if err := i.Validate(); err != nil {
		return err
	}
	var rid *string
	g1pk, g1sk := incidentByServiceGSI1(i.ServiceID, i.OpenedAt)
	g2pk, g2sk := "", ""
	if i.ReleaseID != nil {
		v := i.ReleaseID.String()
		rid = &v
		g2pk, g2sk = incidentByReleaseGSI2(*i.ReleaseID, i.OpenedAt)
	}
	return s.putNew(ctx, "incident", record{
		PK: incidentPK(i.ID), SK: metaSK(), GSI1PK: g1pk, GSI1SK: g1sk, GSI2PK: g2pk, GSI2SK: g2sk,
		EntityType: "INCIDENT", Payload: mustPayload(incidentPayload{
			ID: i.ID.String(), ServiceID: i.ServiceID.String(), ReleaseID: rid, Title: i.Title,
			Status: string(i.Status), OpenedAt: i.OpenedAt, ResolvedAt: i.ResolvedAt, Source: string(i.Source),
		}),
	}, i.ID.String())
}

func (s *Store) CreateEvent(ctx context.Context, e domain.ReleaseEvent) error {
	e = e.Normalized()
	if err := e.Validate(); err != nil {
		return err
	}
	payload := eventPayload{
		ID: e.ID.String(), Type: string(e.Type), SchemaVersion: e.SchemaVersion,
		OccurredAt: e.OccurredAt, IngestedAt: e.IngestedAt, Producer: string(e.Producer),
		CorrelationID: e.CorrelationID,
	}
	if e.ReleaseID != nil {
		v := e.ReleaseID.String()
		payload.ReleaseID = &v
	}
	if e.ServiceID != nil {
		v := e.ServiceID.String()
		payload.ServiceID = &v
	}
	if e.DeploymentID != nil {
		v := e.DeploymentID.String()
		payload.DeploymentID = &v
	}
	raw := mustPayload(payload)
	idem := record{PK: idemPK(e.ID), SK: metaSK(), EntityType: "IDEM", Payload: raw}
	items := []types.TransactWriteItem{s.conditionalPut(idem)}
	if e.ReleaseID != nil {
		items = append(items, s.conditionalPut(record{
			PK: releasePK(*e.ReleaseID), SK: eventSK(e.OccurredAt, e.ID),
			EntityType: "EVENT", Payload: raw,
		}))
	}
	_, err := s.api.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: items})
	if err != nil {
		if isConditional(err) {
			return domain.AlreadyExistsError{Resource: "event", ID: e.ID.String()}
		}
		return wrapErr("create_event", err)
	}
	return nil
}

func (s *Store) GetEvent(ctx context.Context, id domain.EventID) (domain.ReleaseEvent, error) {
	rec, err := s.get(ctx, idemPK(id), metaSK())
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.ReleaseEvent{}, domain.NotFoundError{Resource: "event", ID: id.String()}
		}
		return domain.ReleaseEvent{}, err
	}
	var p eventPayload
	if err := decodePayload(rec.Payload, &p); err != nil {
		return domain.ReleaseEvent{}, wrapErr("get_event", err)
	}
	return eventFrom(p).Normalized(), nil
}

func (s *Store) ListEventsByRelease(ctx context.Context, releaseID domain.ReleaseID) ([]domain.ReleaseEvent, error) {
	items, err := s.queryPK(ctx, releasePK(releaseID), "EVENT#")
	if err != nil {
		return nil, err
	}
	out := make([]domain.ReleaseEvent, 0, len(items))
	for _, rec := range items {
		var p eventPayload
		if err := decodePayload(rec.Payload, &p); err != nil {
			return nil, wrapErr("list_events", err)
		}
		out = append(out, eventFrom(p).Normalized())
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].OccurredAt.Equal(out[j].OccurredAt) {
			return out[i].ID.String() < out[j].ID.String()
		}
		return out[i].OccurredAt.Before(out[j].OccurredAt)
	})
	return out, nil
}

func (s *Store) PutRiskAssessment(ctx context.Context, a domain.RiskAssessment) error {
	raw, err := encodePayload(a)
	if err != nil {
		return wrapErr("put_risk", err)
	}
	item, err := marshalRecord(record{PK: releasePK(a.ReleaseID), SK: "RISK#LATEST", EntityType: "RISK", Payload: raw})
	if err != nil {
		return wrapErr("put_risk", err)
	}
	_, err = s.api.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(s.table), Item: item})
	return wrapErr("put_risk", err)
}

func (s *Store) GetRiskAssessment(ctx context.Context, releaseID domain.ReleaseID) (domain.RiskAssessment, error) {
	rec, err := s.get(ctx, releasePK(releaseID), "RISK#LATEST")
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.RiskAssessment{}, domain.NotFoundError{Resource: "risk", ID: releaseID.String()}
		}
		return domain.RiskAssessment{}, err
	}
	var a domain.RiskAssessment
	if err := decodePayload(rec.Payload, &a); err != nil {
		return domain.RiskAssessment{}, wrapErr("get_risk", err)
	}
	return a, nil
}

func (s *Store) CreateSecurityScan(ctx context.Context, scan domain.SecurityScan) error {
	raw, err := encodePayload(scan)
	if err != nil {
		return wrapErr("put_scan", err)
	}
	return s.putNew(ctx, "security_scan", record{
		PK: releasePK(scan.ReleaseID), SK: "SCAN#" + scan.ID, EntityType: "SCAN", Payload: raw,
	}, scan.ID)
}

func (s *Store) GetSecurityScanByRelease(ctx context.Context, releaseID domain.ReleaseID) (domain.SecurityScan, error) {
	items, err := s.queryPK(ctx, releasePK(releaseID), "SCAN#")
	if err != nil {
		return domain.SecurityScan{}, err
	}
	if len(items) == 0 {
		return domain.SecurityScan{}, domain.NotFoundError{Resource: "security_scan", ID: releaseID.String()}
	}
	var scan domain.SecurityScan
	if err := decodePayload(items[0].Payload, &scan); err != nil {
		return domain.SecurityScan{}, wrapErr("get_scan", err)
	}
	return scan, nil
}

func (s *Store) ListIncidentsByService(ctx context.Context, serviceID domain.ServiceID) ([]domain.Incident, error) {
	items, err := s.queryIndex(ctx, "GSI1", "GSI1PK = :pk AND begins_with(GSI1SK, :sk)", map[string]types.AttributeValue{
		":pk": &types.AttributeValueMemberS{Value: servicePK(serviceID)},
		":sk": &types.AttributeValueMemberS{Value: "INCIDENT#"},
	}, true)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Incident, 0, len(items))
	for _, rec := range items {
		var p incidentPayload
		if err := decodePayload(rec.Payload, &p); err != nil {
			return nil, wrapErr("list_incidents", err)
		}
		var rid *domain.ReleaseID
		if p.ReleaseID != nil {
			id := domain.ReleaseID(*p.ReleaseID)
			rid = &id
		}
		out = append(out, domain.Incident{
			ID: domain.IncidentID(p.ID), ServiceID: domain.ServiceID(p.ServiceID), ReleaseID: rid,
			Title: p.Title, Status: domain.IncidentStatus(p.Status), OpenedAt: p.OpenedAt, ResolvedAt: p.ResolvedAt,
			Source: domain.DataSource(p.Source),
		}.Normalized())
	}
	return out, nil
}

func (s *Store) ListHealthSnapshotsByService(ctx context.Context, serviceID domain.ServiceID) ([]domain.HealthSnapshot, error) {
	items, err := s.queryPK(ctx, servicePK(serviceID), "HEALTH#")
	if err != nil {
		return nil, err
	}
	out := make([]domain.HealthSnapshot, 0, len(items))
	for _, rec := range items {
		var p healthPayload
		if err := decodePayload(rec.Payload, &p); err != nil {
			return nil, wrapErr("list_health", err)
		}
		out = append(out, healthFrom(p).Normalized())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WindowStart.Before(out[j].WindowStart) })
	return out, nil
}

func (s *Store) PutHealthComparison(ctx context.Context, a domain.HealthAssessment) error {
	raw, err := encodePayload(a)
	if err != nil {
		return wrapErr("put_health_cmp", err)
	}
	item, err := marshalRecord(record{PK: releasePK(a.ReleaseID), SK: "HEALTH#COMPARISON", EntityType: "HEALTHCMP", Payload: raw})
	if err != nil {
		return wrapErr("put_health_cmp", err)
	}
	_, err = s.api.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(s.table), Item: item})
	return wrapErr("put_health_cmp", err)
}

func (s *Store) GetHealthComparison(ctx context.Context, releaseID domain.ReleaseID) (domain.HealthAssessment, error) {
	rec, err := s.get(ctx, releasePK(releaseID), "HEALTH#COMPARISON")
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.HealthAssessment{}, domain.NotFoundError{Resource: "health", ID: releaseID.String()}
		}
		return domain.HealthAssessment{}, err
	}
	var a domain.HealthAssessment
	if err := decodePayload(rec.Payload, &a); err != nil {
		return domain.HealthAssessment{}, wrapErr("get_health_cmp", err)
	}
	return a, nil
}

func (s *Store) PutRollbackAssessment(ctx context.Context, a domain.RollbackAssessment) error {
	raw, err := encodePayload(a)
	if err != nil {
		return wrapErr("put_rollback", err)
	}
	item, err := marshalRecord(record{PK: releasePK(a.ReleaseID), SK: "ROLLBACK#LATEST", EntityType: "ROLLBACK", Payload: raw})
	if err != nil {
		return wrapErr("put_rollback", err)
	}
	_, err = s.api.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(s.table), Item: item})
	return wrapErr("put_rollback", err)
}

func (s *Store) GetRollbackAssessment(ctx context.Context, releaseID domain.ReleaseID) (domain.RollbackAssessment, error) {
	rec, err := s.get(ctx, releasePK(releaseID), "ROLLBACK#LATEST")
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.RollbackAssessment{}, domain.NotFoundError{Resource: "rollback", ID: releaseID.String()}
		}
		return domain.RollbackAssessment{}, err
	}
	var a domain.RollbackAssessment
	if err := decodePayload(rec.Payload, &a); err != nil {
		return domain.RollbackAssessment{}, wrapErr("get_rollback", err)
	}
	return a, nil
}

func (s *Store) PutPolicyEvaluation(ctx context.Context, e domain.PolicyEvaluation) error {
	raw, err := encodePayload(e)
	if err != nil {
		return wrapErr("put_policy", err)
	}
	item, err := marshalRecord(record{PK: releasePK(e.ReleaseID), SK: "POLICY#LATEST", EntityType: "POLICY", Payload: raw})
	if err != nil {
		return wrapErr("put_policy", err)
	}
	_, err = s.api.PutItem(ctx, &dynamodb.PutItemInput{TableName: aws.String(s.table), Item: item})
	return wrapErr("put_policy", err)
}

func (s *Store) GetPolicyEvaluation(ctx context.Context, releaseID domain.ReleaseID) (domain.PolicyEvaluation, error) {
	rec, err := s.get(ctx, releasePK(releaseID), "POLICY#LATEST")
	if err != nil {
		if domain.IsNotFound(err) {
			return domain.PolicyEvaluation{}, domain.NotFoundError{Resource: "policy", ID: releaseID.String()}
		}
		return domain.PolicyEvaluation{}, err
	}
	var e domain.PolicyEvaluation
	if err := decodePayload(rec.Payload, &e); err != nil {
		return domain.PolicyEvaluation{}, wrapErr("get_policy", err)
	}
	return e, nil
}

func (s *Store) CreateDecision(ctx context.Context, d domain.ReleaseDecision) error {
	d = d.Normalized()
	if err := d.Validate(); err != nil {
		return err
	}
	return s.putNew(ctx, "decision", record{
		PK: releasePK(d.ReleaseID), SK: decisionSK(d.ID), EntityType: "DECISION",
		Payload: mustPayload(decisionPayload{
			ID: d.ID.String(), ReleaseID: d.ReleaseID.String(), Decision: string(d.Decision),
			Actor: d.Actor, Reason: d.Reason, DecidedAt: d.DecidedAt,
		}),
	}, d.ID.String())
}

func (s *Store) putNew(ctx context.Context, resource string, rec record, id string) error {
	item, err := marshalRecord(rec)
	if err != nil {
		return wrapErr("put_"+resource, err)
	}
	_, err = s.api.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		if isConditional(err) {
			return domain.AlreadyExistsError{Resource: resource, ID: id}
		}
		return wrapErr("put_"+resource, err)
	}
	return nil
}

func (s *Store) conditionalPut(rec record) types.TransactWriteItem {
	item, err := marshalRecord(rec)
	if err != nil {
		panic(err)
	}
	return types.TransactWriteItem{Put: &types.Put{
		TableName:           aws.String(s.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	}}
}

func (s *Store) get(ctx context.Context, pk, sk string) (record, error) {
	out, err := s.api.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return record{}, wrapErr("get", err)
	}
	if out.Item == nil {
		return record{}, domain.ErrNotFound
	}
	return unmarshalRecord(out.Item)
}

func (s *Store) queryPK(ctx context.Context, pk, skPrefix string) ([]record, error) {
	out, err := s.api.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.table),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pk},
			":sk": &types.AttributeValueMemberS{Value: skPrefix},
		},
	})
	if err != nil {
		return nil, wrapErr("query", err)
	}
	return decodeRecords(out.Items)
}

func (s *Store) queryIndex(ctx context.Context, index, expr string, values map[string]types.AttributeValue, forward bool) ([]record, error) {
	out, err := s.api.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(s.table),
		IndexName:                 aws.String(index),
		KeyConditionExpression:    aws.String(expr),
		ExpressionAttributeValues: values,
		ScanIndexForward:          aws.Bool(forward),
	})
	if err != nil {
		return nil, wrapErr("query_index", err)
	}
	return decodeRecords(out.Items)
}

func decodeRecords(items []map[string]types.AttributeValue) ([]record, error) {
	out := make([]record, 0, len(items))
	for _, it := range items {
		rec, err := unmarshalRecord(it)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func mustPayload(v any) string {
	s, err := encodePayload(v)
	if err != nil {
		panic(err)
	}
	return s
}
