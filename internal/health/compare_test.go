package health

import (
	"testing"
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

func healthy() *Window {
	return &Window{RequestCount: 200, ErrorRate: 0.007, Availability: 0.9996, P95: 182, HasError: true, HasAvail: true, HasP95: true}
}

func TestStable(t *testing.T) {
	c := Compare(Input{DeployedAt: time.Now().UTC(), Baseline: healthy(), Post: healthy(), ChangedService: "checkout-api", DegradedService: "checkout-api"})
	if c.Overall != Stable || c.Correlation != NoClearCorrelation {
		t.Fatalf("%+v", c)
	}
}

func TestClearDegradationAndCorrelation(t *testing.T) {
	post := healthy()
	post.ErrorRate = 0.049
	post.Availability = 0.988
	post.P95 = 463
	c := Compare(Input{
		DeployedAt: time.Now().UTC(), Baseline: healthy(), Post: post,
		ChangedService: "checkout-api", DegradedService: "checkout-api",
	})
	if c.Overall != SeverelyDegraded {
		t.Fatalf("overall=%s", c.Overall)
	}
	if c.Correlation != LikelyCorrelation {
		t.Fatalf("corr=%s reasons=%v", c.Correlation, c.Reasons)
	}
}

func TestZeroBaselineNoPct(t *testing.T) {
	b := healthy()
	b.ErrorRate = 0
	p := healthy()
	p.ErrorRate = 0.002
	c := Compare(Input{DeployedAt: time.Now().UTC(), Baseline: b, Post: p})
	for _, m := range c.Metrics {
		if m.Name == "error_rate" && m.PctDelta != nil {
			t.Fatal("pct delta must be omitted when baseline is 0")
		}
	}
}

func TestMissingAndLowSample(t *testing.T) {
	if Compare(Input{DeployedAt: time.Now().UTC()}).Overall != InsufficientData {
		t.Fatal("missing windows")
	}
	low := healthy()
	low.RequestCount = 10
	if Compare(Input{DeployedAt: time.Now().UTC(), Baseline: low, Post: healthy()}).Overall != InsufficientData {
		t.Fatal("low sample")
	}
}

func TestFalseAttribution(t *testing.T) {
	post := healthy()
	post.ErrorRate = 0.04
	c := Compare(Input{
		DeployedAt: time.Now().UTC(), Baseline: healthy(), Post: post,
		ChangedService: "checkout-api", DegradedService: "checkout-api", BaselineAlreadyBad: true,
	})
	if c.Correlation != NoClearCorrelation {
		t.Fatalf("%s", c.Correlation)
	}
	c = Compare(Input{
		DeployedAt: time.Now().UTC(), Baseline: healthy(), Post: post,
		OverlappingDeployment: true, ChangedService: "a", DegradedService: "a",
	})
	if c.Correlation != CorrInsufficient {
		t.Fatalf("%s", c.Correlation)
	}
	c = Compare(Input{
		DeployedAt: time.Now().UTC(), Baseline: healthy(), Post: post,
		DegradationBeforeDeploy: true, ChangedService: "a", DegradedService: "a",
	})
	if c.Correlation != NoClearCorrelation {
		t.Fatalf("%s", c.Correlation)
	}
	_ = domain.ServiceID("x")
}

func TestMinorNoiseStaysStable(t *testing.T) {
	post := healthy()
	post.ErrorRate = 0.0072
	post.P95 = 190
	c := Compare(Input{DeployedAt: time.Now().UTC(), Baseline: healthy(), Post: post})
	if c.Overall != Stable {
		t.Fatalf("%s", c.Overall)
	}
}

func TestPartialMetricsInsufficientBeatsStable(t *testing.T) {
	b := healthy()
	b.HasP95 = false
	p := healthy()
	p.HasP95 = false
	c := Compare(Input{DeployedAt: time.Now().UTC(), Baseline: b, Post: p})
	if c.Overall != InsufficientData {
		t.Fatalf("partial metrics must not silently become STABLE, got %s", c.Overall)
	}
	for _, m := range c.Metrics {
		if m.Name == "latency_p95_ms" && m.Available {
			t.Fatal("p95 should be unavailable")
		}
	}
}

func TestDeterminism(t *testing.T) {
	in := Input{DeployedAt: time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC), Baseline: healthy(), Post: healthy(), ChangedService: "a", DegradedService: "a"}
	a := Compare(in)
	b := Compare(in)
	if a.Overall != b.Overall || a.Correlation != b.Correlation {
		t.Fatal("not deterministic")
	}
}
