package risk

import (
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

func ptr[T any](v T) *T { return &v }

func TestDeterminismAndClipping(t *testing.T) {
	in := Input{
		ServiceCriticality:    domain.CriticalityCritical,
		FilesChanged:          ptr(80),
		LinesChanged:          ptr(2000),
		MigrationPresent:      ptr(true),
		ConfigChangePresent:   ptr(true),
		FailedCIAttempts:      ptr(3),
		FailedTests:           ptr(2),
		RecentDeployFailures:  ptr(2),
		DependentCount:        ptr(8),
		RecentIncidents:       ptr(3),
		Rollback:              ptr(domain.RollbackNotReady),
		HighestSeverity:       ptr(domain.SeverityCritical),
		HistoricalFailureRate: ptr(0.4),
		Now:                   time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC),
	}
	a := Calculate("rel_x", in)
	b := Calculate("rel_x", in)
	if a.ScoreRaw != b.ScoreRaw || a.Score != b.Score {
		t.Fatalf("not deterministic %d vs %d", a.ScoreRaw, b.ScoreRaw)
	}
	if a.ScoreRaw <= 100 {
		t.Fatalf("expected raw over 100, got %d", a.ScoreRaw)
	}
	if a.Score != 100 || a.Category != domain.RiskCritical {
		t.Fatalf("clip=%d cat=%s", a.Score, a.Category)
	}
	sum := 0
	for _, f := range a.Factors {
		sum += f.Points
	}
	if sum != a.ScoreRaw {
		t.Fatalf("factor sum %d != raw %d", sum, a.ScoreRaw)
	}
}

func TestMissingDataAwardsNoPoints(t *testing.T) {
	a := Calculate("rel_y", Input{ServiceCriticality: domain.CriticalityLow, Now: time.Now().UTC()})
	if a.Score != 0 || a.Category != domain.RiskLow {
		t.Fatalf("score=%d cat=%s", a.Score, a.Category)
	}
	omitted := 0
	for _, f := range a.Factors {
		if f.Omitted && f.Points != 0 {
			t.Fatalf("omitted factor scored: %+v", f)
		}
		if f.Omitted {
			omitted++
		}
	}
	if omitted < 8 {
		t.Fatalf("expected many omitted factors, got %d", omitted)
	}
}

func TestCategoryBoundaries(t *testing.T) {
	if domain.CategoryForScore(24) != domain.RiskLow || domain.CategoryForScore(25) != domain.RiskModerate {
		t.Fatal("24/25")
	}
	if domain.CategoryForScore(49) != domain.RiskModerate || domain.CategoryForScore(50) != domain.RiskHigh {
		t.Fatal("49/50")
	}
	if domain.CategoryForScore(74) != domain.RiskHigh || domain.CategoryForScore(75) != domain.RiskCritical {
		t.Fatal("74/75")
	}
}

func TestUnknownServiceDefaultsHigh(t *testing.T) {
	a := Calculate("rel_z", Input{UnknownService: true, Now: time.Now().UTC()})
	var crit domain.RiskFactor
	for _, f := range a.Factors {
		if f.Code == "service_criticality" {
			crit = f
		}
	}
	if crit.Points != 10 || crit.Input != string(domain.CriticalityHigh) {
		t.Fatalf("%+v", crit)
	}
}

func TestChangeSizeBands(t *testing.T) {
	a := Calculate("rel", Input{FilesChanged: ptr(6), LinesChanged: ptr(10), ServiceCriticality: domain.CriticalityLow, Now: time.Now().UTC()})
	got := factor(a, "change_size").Points
	if got != 3 {
		t.Fatalf("files band got %d", got)
	}
	b := Calculate("rel", Input{FilesChanged: ptr(1), LinesChanged: ptr(400), ServiceCriticality: domain.CriticalityLow, Now: time.Now().UTC()})
	if factor(b, "change_size").Points != 7 {
		t.Fatalf("line band")
	}
}

func factor(a domain.RiskAssessment, code string) domain.RiskFactor {
	for _, f := range a.Factors {
		if f.Code == code {
			return f
		}
	}
	return domain.RiskFactor{}
}
