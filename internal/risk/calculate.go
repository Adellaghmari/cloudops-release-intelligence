package risk

import (
	"fmt"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
)

const ModelVersion = "risk-v1"

type Input struct {
	ServiceCriticality    domain.Criticality
	UnknownService        bool
	FilesChanged          *int
	LinesChanged          *int
	MigrationPresent      *bool
	ConfigChangePresent   *bool
	FailedCIAttempts      *int
	FailedTests           *int
	RecentDeployFailures  *int
	DependentCount        *int
	RecentIncidents       *int
	Rollback              *domain.RollbackStatus
	HighestSeverity       *domain.SecuritySeverity
	HistoricalFailureRate *float64
	Now                   time.Time
}

func Calculate(releaseID domain.ReleaseID, in Input) domain.RiskAssessment {
	now := in.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var factors []domain.RiskFactor
	factors = append(factors, changeSize(in))
	factors = append(factors, criticality(in))
	factors = append(factors, flagFactor("database_migration", "Database migration", 15, in.MigrationPresent, "migration present"))
	factors = append(factors, flagFactor("config_change", "Configuration change", 8, in.ConfigChangePresent, "config change present"))
	factors = append(factors, ciAttempts(in))
	factors = append(factors, testFailures(in))
	factors = append(factors, recentFailures(in))
	factors = append(factors, fanout(in))
	factors = append(factors, incidents(in))
	factors = append(factors, rollbackGap(in))
	factors = append(factors, security(in))
	factors = append(factors, historical(in))

	raw := 0
	for _, f := range factors {
		raw += f.Points
	}
	score := raw
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return domain.RiskAssessment{
		ReleaseID:    releaseID,
		Score:        score,
		ScoreRaw:     raw,
		Category:     domain.CategoryForScore(score),
		Factors:      factors,
		ModelVersion: ModelVersion,
		AssessedAt:   now,
	}
}

func changeSize(in Input) domain.RiskFactor {
	if in.FilesChanged == nil && in.LinesChanged == nil {
		return omit("change_size", "Change size", "commit change size not recorded")
	}
	files, lines := 0, 0
	if in.FilesChanged != nil {
		files = *in.FilesChanged
	}
	if in.LinesChanged != nil {
		lines = *in.LinesChanged
	}
	fp, lp := 0, 0
	switch {
	case files >= 51:
		fp = 10
	case files >= 21:
		fp = 6
	case files >= 6:
		fp = 3
	}
	switch {
	case lines >= 1001:
		lp = 10
	case lines >= 301:
		lp = 7
	case lines >= 50:
		lp = 4
	}
	pts := fp
	if lp > pts {
		pts = lp
	}
	return domain.RiskFactor{
		Code: "change_size", Label: "Change size", Points: pts,
		Rationale: "Larger diffs are harder to review. This is a size heuristic, not code quality.",
		Input:     fmt.Sprintf("files=%d lines=%d", files, lines),
	}
}

func criticality(in Input) domain.RiskFactor {
	crit := in.ServiceCriticality
	if in.UnknownService {
		crit = domain.CriticalityHigh
	}
	pts := 0
	switch crit {
	case domain.CriticalityCritical:
		pts = 20
	case domain.CriticalityHigh:
		pts = 10
	case domain.CriticalityModerate:
		pts = 4
	}
	rationale := "A change on a more critical service is a larger operational bet."
	if in.UnknownService {
		rationale += " Unknown service defaulted to HIGH."
	}
	return domain.RiskFactor{Code: "service_criticality", Label: "Service criticality", Points: pts, Rationale: rationale, Input: string(crit)}
}

func flagFactor(code, label string, pts int, flag *bool, whenTrue string) domain.RiskFactor {
	if flag == nil {
		return omit(code, label, "flag not recorded")
	}
	p := 0
	if *flag {
		p = pts
	}
	return domain.RiskFactor{Code: code, Label: label, Points: p, Rationale: "Presence is the signal; quality of the change is not inferred.", Input: fmt.Sprintf("%s=%t", whenTrue, *flag)}
}

func ciAttempts(in Input) domain.RiskFactor {
	if in.FailedCIAttempts == nil {
		return omit("ci_failed_attempts", "Failed CI attempts", "CI attempt count not recorded")
	}
	pts := *in.FailedCIAttempts * 4
	if pts > 10 {
		pts = 10
	}
	return domain.RiskFactor{Code: "ci_failed_attempts", Label: "Failed CI attempts", Points: pts, Rationale: "Repeated CI failure is process risk, not proof the final artifact is bad.", Input: fmt.Sprintf("failed_attempts=%d", *in.FailedCIAttempts)}
}

func testFailures(in Input) domain.RiskFactor {
	if in.FailedTests == nil {
		return omit("test_failures", "Test failures", "test results not recorded")
	}
	pts := 0
	if *in.FailedTests > 0 {
		pts = 15
	}
	return domain.RiskFactor{Code: "test_failures", Label: "Test failures", Points: pts, Rationale: "Failing tests on the attached run are a first-class risk signal.", Input: fmt.Sprintf("failed_tests=%d", *in.FailedTests)}
}

func recentFailures(in Input) domain.RiskFactor {
	if in.RecentDeployFailures == nil {
		return omit("recent_deploy_failures", "Recent deployment failures", "deployment history insufficient")
	}
	pts := 0
	switch {
	case *in.RecentDeployFailures >= 2:
		pts = 10
	case *in.RecentDeployFailures == 1:
		pts = 5
	}
	return domain.RiskFactor{Code: "recent_deploy_failures", Label: "Recent deployment failures", Points: pts, Rationale: "Recent failed deploys on the same service raise operational caution.", Input: fmt.Sprintf("failed_14d=%d", *in.RecentDeployFailures)}
}

func fanout(in Input) domain.RiskFactor {
	if in.DependentCount == nil {
		return omit("dependency_fanout", "Dependency fan-out", "dependency graph not loaded")
	}
	n := *in.DependentCount
	pts := 0
	switch {
	case n >= 7:
		pts = 12
	case n >= 4:
		pts = 9
	case n >= 2:
		pts = 6
	}
	return domain.RiskFactor{Code: "dependency_fanout", Label: "Dependency fan-out", Points: pts, Rationale: "More dependents increase potential blast radius. This is potential impact, not observed impact.", Input: fmt.Sprintf("dependents=%d", n)}
}

func incidents(in Input) domain.RiskFactor {
	if in.RecentIncidents == nil {
		return omit("recent_incidents", "Recent incidents", "incident history not loaded")
	}
	pts := 0
	switch {
	case *in.RecentIncidents >= 2:
		pts = 8
	case *in.RecentIncidents == 1:
		pts = 4
	}
	return domain.RiskFactor{Code: "recent_incidents", Label: "Recent incidents", Points: pts, Rationale: "Recent incidents on the changed service are a stability signal.", Input: fmt.Sprintf("incidents_30d=%d", *in.RecentIncidents)}
}

func rollbackGap(in Input) domain.RiskFactor {
	if in.Rollback == nil {
		return omit("rollback_gap", "Rollback gap", "rollback readiness not assessed")
	}
	pts := 0
	switch *in.Rollback {
	case domain.RollbackNotReady:
		pts = 7
	case domain.RollbackUnknown:
		pts = 4
	case domain.RollbackPartial:
		pts = 3
	}
	return domain.RiskFactor{Code: "rollback_gap", Label: "Rollback gap", Points: pts, Rationale: "Missing rollback prerequisites increase recovery risk.", Input: string(*in.Rollback)}
}

func security(in Input) domain.RiskFactor {
	if in.HighestSeverity == nil {
		return omit("security_findings", "Security findings", "security scan not attached")
	}
	pts := 0
	switch *in.HighestSeverity {
	case domain.SeverityCritical:
		pts = 15
	case domain.SeverityHigh:
		pts = 10
	case domain.SeverityMedium:
		pts = 4
	}
	return domain.RiskFactor{Code: "security_findings", Label: "Security findings", Points: pts, Rationale: "Highest attached scan severity. Not a complete CVE analysis.", Input: string(*in.HighestSeverity)}
}

func historical(in Input) domain.RiskFactor {
	if in.HistoricalFailureRate == nil {
		return omit("historical_failure_rate", "Historical release failure rate", "fewer than 3 completed deployments")
	}
	rate := *in.HistoricalFailureRate
	pts := 0
	switch {
	case rate >= 0.30:
		pts = 8
	case rate >= 0.10:
		pts = 4
	}
	return domain.RiskFactor{Code: "historical_failure_rate", Label: "Historical release failure rate", Points: pts, Rationale: "Unstable on tiny samples; omitted below 3 deploys.", Input: fmt.Sprintf("failure_rate=%.2f", rate)}
}

func omit(code, label, why string) domain.RiskFactor {
	return domain.RiskFactor{Code: code, Label: label, Points: 0, Omitted: true, Rationale: "No points awarded for unavailable data.", Input: why}
}
