package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"
)

// Typed identifiers. These are first-class domain values, not DynamoDB keys
// and not HTTP path decorations. Persistence adapters map them later.

type ServiceID string
type ReleaseID string
type DeploymentID string
type EventID string
type IncidentID string
type CIRunID string
type HealthSnapshotID string
type DecisionID string
type CommitSHA string

const (
	maxIDLen      = 128
	minIDLen      = 1
	maxSHALen     = 40
	minSHALen     = 7
	generatedSize = 10 // 20 hex chars after prefix
)

func NewReleaseID() ReleaseID       { return ReleaseID("rel_" + randomHex(generatedSize)) }
func NewDeploymentID() DeploymentID { return DeploymentID("dep_" + randomHex(generatedSize)) }
func NewEventID() EventID           { return EventID("evt_" + randomHex(generatedSize)) }
func NewIncidentID() IncidentID     { return IncidentID("inc_" + randomHex(generatedSize)) }
func NewCIRunID() CIRunID           { return CIRunID("ci_" + randomHex(generatedSize)) }
func NewHealthSnapshotID() HealthSnapshotID {
	return HealthSnapshotID("hlt_" + randomHex(generatedSize))
}
func NewDecisionID() DecisionID { return DecisionID("dec_" + randomHex(generatedSize)) }

func ParseServiceID(raw string) (ServiceID, error) {
	id := ServiceID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseReleaseID(raw string) (ReleaseID, error) {
	id := ReleaseID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseDeploymentID(raw string) (DeploymentID, error) {
	id := DeploymentID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseEventID(raw string) (EventID, error) {
	id := EventID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseIncidentID(raw string) (IncidentID, error) {
	id := IncidentID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseCIRunID(raw string) (CIRunID, error) {
	id := CIRunID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseHealthSnapshotID(raw string) (HealthSnapshotID, error) {
	id := HealthSnapshotID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseDecisionID(raw string) (DecisionID, error) {
	id := DecisionID(strings.TrimSpace(raw))
	if err := id.Validate(); err != nil {
		return "", err
	}
	return id, nil
}

func ParseCommitSHA(raw string) (CommitSHA, error) {
	sha := CommitSHA(strings.TrimSpace(strings.ToLower(raw)))
	if err := sha.Validate(); err != nil {
		return "", err
	}
	return sha, nil
}

func (id ServiceID) Validate() error        { return validateID("service_id", string(id)) }
func (id ReleaseID) Validate() error        { return validateID("release_id", string(id)) }
func (id DeploymentID) Validate() error     { return validateID("deployment_id", string(id)) }
func (id EventID) Validate() error          { return validateID("event_id", string(id)) }
func (id IncidentID) Validate() error       { return validateID("incident_id", string(id)) }
func (id CIRunID) Validate() error          { return validateID("ci_run_id", string(id)) }
func (id HealthSnapshotID) Validate() error { return validateID("health_snapshot_id", string(id)) }
func (id DecisionID) Validate() error       { return validateID("decision_id", string(id)) }

func (id ServiceID) String() string        { return string(id) }
func (id ReleaseID) String() string        { return string(id) }
func (id DeploymentID) String() string     { return string(id) }
func (id EventID) String() string          { return string(id) }
func (id IncidentID) String() string       { return string(id) }
func (id CIRunID) String() string          { return string(id) }
func (id HealthSnapshotID) String() string { return string(id) }
func (id DecisionID) String() string       { return string(id) }
func (s CommitSHA) String() string         { return string(s) }

func (s CommitSHA) Validate() error {
	raw := string(s)
	if len(raw) < minSHALen || len(raw) > maxSHALen {
		return ValidationError{Field: "commit_sha", Message: "must be 7-40 hex characters"}
	}
	for _, r := range raw {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return ValidationError{Field: "commit_sha", Message: "must be hexadecimal"}
		}
	}
	return nil
}

func validateID(field, raw string) error {
	if len(raw) < minIDLen || len(raw) > maxIDLen {
		return ValidationError{Field: field, Message: fmt.Sprintf("must be %d-%d characters", minIDLen, maxIDLen)}
	}
	for _, r := range raw {
		if !isIDRune(r) {
			return ValidationError{Field: field, Message: "contains invalid characters"}
		}
	}
	return nil
}

func isIDRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-' || r == ':'
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}

func clonePtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
