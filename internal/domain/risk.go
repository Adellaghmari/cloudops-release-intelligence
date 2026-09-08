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
