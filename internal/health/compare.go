package health

import (
	"math"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

const ModelVersion = "health-v1"

type Verdict string

const (
	Stable           Verdict = "STABLE"
	Degraded         Verdict = "DEGRADED"
	SeverelyDegraded Verdict = "SEVERELY_DEGRADED"
	InsufficientData Verdict = "INSUFFICIENT_DATA"
)

type Correlation string

const (
	LikelyCorrelation   Correlation = "LIKELY_RELEASE_CORRELATION"
	PossibleCorrelation Correlation = "POSSIBLE_RELEASE_CORRELATION"
	NoClearCorrelation  Correlation = "NO_CLEAR_RELEASE_CORRELATION"
	CorrInsufficient    Correlation = "INSUFFICIENT_DATA"
)

type MetricComparison struct {
	Name      string
	Baseline  float64
	Post      float64
	AbsDelta  float64
	PctDelta  *float64
	Threshold string
	Verdict   Verdict
	Available bool
	Reason    string
}

type Comparison struct {
	Overall      Verdict
	Metrics      []MetricComparison
	Correlation  Correlation
	Reasons      []string
	BaselineFrom time.Time
	BaselineTo   time.Time
	PostFrom     time.Time
	PostTo       time.Time
	ModelVersion string
}

type Window struct {
	Start, End    time.Time
	RequestCount  int
	ErrorRate     float64
	Availability  float64
	P50, P95, P99 float64
	HasError      bool
	HasAvail      bool
	HasP95        bool
}

type Input struct {
	DeployedAt              time.Time
	Baseline                *Window
	Post                    *Window
	ChangedService          domain.ServiceID
	DegradedService         domain.ServiceID
	ChangedIsUpstream       bool
	BaselineAlreadyBad      bool
	OverlappingDeployment   bool
	DegradationBeforeDeploy bool
	UnrelatedService        bool
	IncidentInWindow        bool
}

func Compare(in Input) Comparison {
	t := in.DeployedAt.UTC()
	out := Comparison{
		ModelVersion: ModelVersion,
		BaselineFrom: t.Add(-60 * time.Minute),
		BaselineTo:   t,
		PostFrom:     t.Add(2 * time.Minute),
		PostTo:       t.Add(32 * time.Minute),
	}
	if in.Baseline == nil || in.Post == nil {
		out.Overall = InsufficientData
		out.Correlation = CorrInsufficient
		out.Reasons = []string{"missing baseline or post window"}
		return out
	}
	if in.Baseline.RequestCount < 50 || in.Post.RequestCount < 50 {
		out.Overall = InsufficientData
		out.Correlation = CorrInsufficient
		out.Reasons = []string{"request_count below 50 in a window"}
		return out
	}
	out.Metrics = append(out.Metrics, compareError(in.Baseline, in.Post))
	out.Metrics = append(out.Metrics, compareAvail(in.Baseline, in.Post))
	out.Metrics = append(out.Metrics, compareP95(in.Baseline, in.Post))
	out.Overall = worst(out.Metrics)
	out.Correlation, out.Reasons = correlate(in, out.Overall)
	return out
}

func compareError(b, p *Window) MetricComparison {
	m := MetricComparison{Name: "error_rate", Baseline: b.ErrorRate, Post: p.ErrorRate, AbsDelta: p.ErrorRate - b.ErrorRate, Available: b.HasError && p.HasError, Threshold: "2x and +0.5pp or SLO 1%"}
	m.PctDelta = pct(b.ErrorRate, p.ErrorRate)
	if !m.Available {
		m.Verdict = InsufficientData
		m.Reason = "error_rate not present on both windows"
		return m
	}
	if p.ErrorRate >= 0.05 || (p.ErrorRate >= b.ErrorRate*5 && p.ErrorRate >= 0.02) {
		m.Verdict = SeverelyDegraded
		m.Reason = "severe error-rate increase"
		return m
	}
	if p.ErrorRate > math.Max(b.ErrorRate*2, b.ErrorRate+0.005) || p.ErrorRate > 0.01 {
		m.Verdict = Degraded
		m.Reason = "error rate crossed degradation threshold"
		return m
	}
	m.Verdict = Stable
	m.Reason = "within noise"
	return m
}

func compareAvail(b, p *Window) MetricComparison {
	m := MetricComparison{Name: "availability", Baseline: b.Availability, Post: p.Availability, AbsDelta: p.Availability - b.Availability, Available: b.HasAvail && p.HasAvail, Threshold: "drop >0.20pp or below 99.9%"}
	m.PctDelta = pct(b.Availability, p.Availability)
	if !m.Available {
		m.Verdict = InsufficientData
		m.Reason = "availability not present"
		return m
	}
	drop := b.Availability - p.Availability
	if drop > 0.01 || p.Availability < 0.99 {
		m.Verdict = SeverelyDegraded
		m.Reason = "availability drop exceeds 1pp or below 99%"
		return m
	}
	if drop > 0.002 || p.Availability < 0.999 {
		m.Verdict = Degraded
		m.Reason = "availability below SLO or dropped >0.20pp"
		return m
	}
	m.Verdict = Stable
	m.Reason = "within SLO"
	return m
}

func compareP95(b, p *Window) MetricComparison {
	m := MetricComparison{Name: "latency_p95_ms", Baseline: b.P95, Post: p.P95, AbsDelta: p.P95 - b.P95, Available: b.HasP95 && p.HasP95, Threshold: "1.5x and +50ms"}
	m.PctDelta = pct(b.P95, p.P95)
	if !m.Available {
		m.Verdict = InsufficientData
		m.Reason = "p95 not present"
		return m
	}
	if p.P95 > b.P95*3 && p.P95-b.P95 > 200 {
		m.Verdict = SeverelyDegraded
		m.Reason = "p95 more than 3x with +200ms"
		return m
	}
	if p.P95 > b.P95*1.5 && p.P95-b.P95 > 50 {
		m.Verdict = Degraded
		m.Reason = "p95 relative and absolute gates fired"
		return m
	}
	m.Verdict = Stable
	m.Reason = "within latency noise"
	return m
}

func correlate(in Input, overall Verdict) (Correlation, []string) {
	var reasons []string
	if overall == InsufficientData {
		return CorrInsufficient, []string{"health insufficient"}
	}
	if overall == Stable {
		return NoClearCorrelation, []string{"post-deploy health is STABLE"}
	}
	if in.DegradationBeforeDeploy {
		return NoClearCorrelation, []string{"degradation began before deployment"}
	}
	if in.BaselineAlreadyBad {
		return NoClearCorrelation, []string{"baseline already breached SLO"}
	}
	if in.OverlappingDeployment {
		return CorrInsufficient, []string{"overlapping neighboring deployment"}
	}
	if in.UnrelatedService {
		return NoClearCorrelation, []string{"degraded service is unrelated"}
	}
	match := in.ChangedService != "" && in.ChangedService == in.DegradedService
	if match {
		reasons = append(reasons, "degraded service is the changed service")
	}
	if in.ChangedIsUpstream {
		reasons = append(reasons, "changed service is upstream of degraded service")
	}
	if in.IncidentInWindow {
		reasons = append(reasons, "incident opened in post window")
	}
	if (match || in.ChangedIsUpstream) && !in.BaselineAlreadyBad {
		if overall == SeverelyDegraded {
			return LikelyCorrelation, append(reasons, "severe post-deploy regression with temporal proximity")
		}
		return PossibleCorrelation, append(reasons, "degradation after deploy with service relationship")
	}
	return CorrInsufficient, []string{"no service or graph relationship"}
}

func worst(ms []MetricComparison) Verdict {
	rank := map[Verdict]int{Stable: 0, InsufficientData: 1, Degraded: 2, SeverelyDegraded: 3}
	best := Stable
	for _, m := range ms {
		if !m.Available {
			if rank[InsufficientData] > rank[best] {
				best = InsufficientData
			}
			continue
		}
		if rank[m.Verdict] > rank[best] {
			best = m.Verdict
		}
	}
	return best
}

func pct(base, post float64) *float64 {
	if base == 0 {
		return nil
	}
	v := (post - base) / base * 100
	return &v
}
