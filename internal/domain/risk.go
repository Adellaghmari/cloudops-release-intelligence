package domain

import "time"

type RiskCategory string

const (
	RiskLow      RiskCategory = "LOW"
	RiskModerate RiskCategory = "MODERATE"
	RiskHigh     RiskCategory = "HIGH"
	RiskCritical RiskCategory = "CRITICAL"
)

type SecuritySeverity string

const (
	SeverityNone     SecuritySeverity = "NONE"
	SeverityMedium   SecuritySeverity = "MEDIUM"
	SeverityHigh     SecuritySeverity = "HIGH"
	SeverityCritical SecuritySeverity = "CRITICAL"
)

type RollbackStatus string

const (
	RollbackReady    RollbackStatus = "READY"
	RollbackPartial  RollbackStatus = "PARTIAL"
	RollbackNotReady RollbackStatus = "NOT_READY"
	RollbackUnknown  RollbackStatus = "UNKNOWN"
)

type RiskFactor struct {
	Code      string
	Label     string
	Points    int
	Rationale string
	Input     string
	Omitted   bool
}

type RiskAssessment struct {
	ReleaseID    ReleaseID
	Score        int
	ScoreRaw     int
	Category     RiskCategory
	Factors      []RiskFactor
	ModelVersion string
	AssessedAt   time.Time
}

type SecurityScan struct {
	ID              string
	ReleaseID       ReleaseID
	HighestSeverity SecuritySeverity
	FindingCount    int
	ScannedAt       time.Time
}

func ParseSecuritySeverity(raw string) (SecuritySeverity, error) {
	switch SecuritySeverity(raw) {
	case SeverityNone, SeverityMedium, SeverityHigh, SeverityCritical:
		return SecuritySeverity(raw), nil
	default:
		return "", ValidationError{Field: "highest_severity", Message: "invalid severity"}
	}
}

type HealthMetricResult struct {
	Name      string
	Baseline  float64
	Post      float64
	AbsDelta  float64
	PctDelta  *float64
	Threshold string
	Verdict   string
	Available bool
	Reason    string
}

type HealthAssessment struct {
	ReleaseID    ReleaseID
	Overall      string
	Correlation  string
	Reasons      []string
	Metrics      []HealthMetricResult
	BaselineFrom time.Time
	BaselineTo   time.Time
	PostFrom     time.Time
	PostTo       time.Time
	ModelVersion string
	ComparedAt   time.Time
	Disclaimer   string
}

func CategoryForScore(score int) RiskCategory {
	switch {
	case score >= 75:
		return RiskCritical
	case score >= 50:
		return RiskHigh
	case score >= 25:
		return RiskModerate
	default:
		return RiskLow
	}
}
