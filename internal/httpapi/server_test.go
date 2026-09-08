package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/config"
	"github.com/adell/cloudops-release-intelligence/internal/domain"
	"github.com/adell/cloudops-release-intelligence/internal/events"
	"github.com/adell/cloudops-release-intelligence/internal/localseed"
	"github.com/adell/cloudops-release-intelligence/internal/repository/memory"
	"github.com/adell/cloudops-release-intelligence/internal/service"
	"github.com/gin-gonic/gin"
)

func testServer(t *testing.T) http.Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	store := memory.New()
	if err := localseed.Load(t.Context(), store, time.Date(2026, 9, 8, 21, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Env:         "local",
		Version:     "0.1.0-test",
		ServiceName: "cloudops-api",
		CORSOrigins: []string{"http://localhost:4200"},
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return NewEngine(cfg, service.NewCatalog(store), logger, events.NewProcessor(store, 3))
}

func TestIngestEventIdempotent(t *testing.T) {
	h := testServer(t)
	body := `{"event_id":"evt_api_1","event_type":"ci.run.completed","occurred_at":"2026-09-08T12:00:00Z","source":"cloudops-api","schema_version":"1.0","release_id":"rel_northstar_payments_demo","correlation_id":"corr_api_1"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("duplicate status=%d body=%s", rec.Code, rec.Body.String())
	}
	var resp eventIngestResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Duplicate {
		t.Fatalf("%+v", resp)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"event_id":"bad id","event_type":"ci.run.completed","occurred_at":"2026-09-08T12:00:00Z","source":"cloudops-api","schema_version":"1.0"}`))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("malformed status=%d", rec.Code)
	}
	if contains(rec.Body.String(), "ValidationException") || contains(rec.Body.String(), "aws") {
		t.Fatalf("leaked aws error: %s", rec.Body.String())
	}
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" || body.Service != "cloudops-api" || body.Version != "0.1.0-test" {
		t.Fatalf("unexpected health: %+v", body)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected generated request id")
	}
}

func TestReadyDoesNotClaimAWS(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	raw := rec.Body.String()
	if contains(raw, "dynamodb") || contains(raw, "DynamoDB") || contains(raw, "lambda") {
		t.Fatalf("readiness must not pretend AWS exists: %s", raw)
	}
	var body readyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Dependencies) != 1 || body.Dependencies[0].Name != "in_memory_store" {
		t.Fatalf("deps=%v", body.Dependencies)
	}
}

func TestRequestIDEchoAndInvalidIgnored(t *testing.T) {
	h := testServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "req_incoming_123")
	req.Header.Set("X-Correlation-ID", "corr_incoming_123")
	h.ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-ID") != "req_incoming_123" {
		t.Fatalf("echo=%s", rec.Header().Get("X-Request-ID"))
	}
	if rec.Header().Get("X-Correlation-ID") != "corr_incoming_123" {
		t.Fatalf("corr=%s", rec.Header().Get("X-Correlation-ID"))
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "has spaces and /slash")
	h.ServeHTTP(rec, req)
	got := rec.Header().Get("X-Request-ID")
	if got == "" || got == "has spaces and /slash" {
		t.Fatalf("invalid header should be replaced, got %q", got)
	}
}

func TestPublicErrorContract(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services/bad%20id", nil)
	req.Header.Set("X-Request-ID", "req_err_1")
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
	var body apiErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "INVALID_ID" || body.Error.RequestID != "req_err_1" || body.Error.Message == "" {
		t.Fatalf("error body=%+v", body.Error)
	}
	if contains(rec.Body.String(), "panic") || contains(rec.Body.String(), "goroutine") || contains(rec.Body.String(), "internal/domain") {
		t.Fatalf("leaked internals: %s", rec.Body.String())
	}
}

func TestNotFoundAndSourceFilter(t *testing.T) {
	h := testServer(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services/missing-service", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var errBody apiErrorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &errBody)
	if errBody.Error.Code != "SERVICE_NOT_FOUND" {
		t.Fatalf("code=%s", errBody.Error.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/releases/missing-release", nil)
	h.ServeHTTP(rec, req)
	_ = json.Unmarshal(rec.Body.Bytes(), &errBody)
	if rec.Code != 404 || errBody.Error.Code != "RELEASE_NOT_FOUND" {
		t.Fatalf("release not found contract: %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/services?source=not-a-source", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/services?source=live", nil)
	h.ServeHTTP(rec, req)
	var list serviceListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if len(list.Services) == 0 {
		t.Fatal("expected live catalog identities")
	}
	for _, s := range list.Services {
		if s.Source != string(domain.DataSourceLive) {
			t.Fatalf("live filter leaked %s", s.Source)
		}
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/releases?source=live", nil)
	h.ServeHTTP(rec, req)
	var rels releaseListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &rels); err != nil {
		t.Fatal(err)
	}
	if len(rels.Releases) != 0 {
		t.Fatalf("must not invent live releases: %+v", rels.Releases)
	}
}

func TestTimelineAndReplay(t *testing.T) {
	h := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/rel_northstar_payments_demo/timeline", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var tl timelineResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &tl); err != nil {
		t.Fatal(err)
	}
	if len(tl.Entries) < 2 {
		t.Fatalf("timeline=%+v", tl)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/replay?a=rel_northstar_payments_demo&b=rel_northstar_regression", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var rp replayResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &rp); err != nil {
		t.Fatal(err)
	}
	if len(rp.Fields) == 0 {
		t.Fatal("expected replay fields")
	}
}

func TestReleaseRollback(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/rel_northstar_payments_demo/rollback", nil)
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var body rollbackResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status == "" || body.Disclaimer == "" {
		t.Fatalf("%+v", body)
	}
	if contains(strings.ToLower(body.Disclaimer), "will roll back") {
		t.Fatal(body.Disclaimer)
	}
}

func TestReleasePolicy(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/rel_northstar_payments_demo/policy", nil)
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var body policyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Result == "" || body.PolicyVersion == "" || len(body.Rules) == 0 {
		t.Fatalf("%+v", body)
	}
}

func TestReleaseImpact(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/rel_northstar_payments_demo/impact", nil)
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var body impactResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ChangedService != "payments-service" || len(body.DirectDependents) == 0 {
		t.Fatalf("%+v", body)
	}
	if contains(strings.ToLower(body.Disclaimer), "definitely") {
		t.Fatal(body.Disclaimer)
	}
}

func TestReleaseHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/releases/rel_northstar_payments_demo/health", nil)
	testServer(t).ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var body healthCompareResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Overall == "" || body.Correlation == "" || body.Disclaimer == "" {
		t.Fatalf("%+v", body)
	}
	if contains(body.Correlation, "PROVEN") || contains(strings.ToLower(body.Disclaimer), "caused") {
		t.Fatalf("must not claim causation: %+v", body)
	}
}

func TestServiceAndReleaseDetail(t *testing.T) {
	h := testServer(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/services/payments-service", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var svc serviceDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &svc); err != nil {
		t.Fatal(err)
	}
	if svc.Service.Source != "synthetic" {
		t.Fatalf("source=%s", svc.Service.Source)
	}
	if len(svc.DependsOn) == 0 || len(svc.DependedBy) == 0 {
		t.Fatalf("graph edges missing: %+v", svc)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/releases/rel_northstar_payments_demo", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	var rel releaseDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &rel); err != nil {
		t.Fatal(err)
	}
	if rel.Release.Source != "synthetic" || rel.Commit == nil || rel.Deployment == nil || rel.CIRun == nil {
		t.Fatalf("detail incomplete: %+v", rel)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
