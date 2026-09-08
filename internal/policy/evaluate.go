package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"
	"github.com/Adellaghmari/cloudops-release-intelligence/policies"
	"github.com/open-policy-agent/opa/v1/rego"
)

const engineRule = "engine.evaluation_failure"

var knownRules = []string{
	"tests.must_pass",
	"security.critical_vuln",
	"security.high_vuln",
	"risk.high_threshold",
	"risk.critical_threshold",
	"migration.rollback_plan",
	"critical_service.change",
	"health.availability_slo",
	"health.severe_regression",
}

type Input struct {
	ReleaseID     string
	ServiceID     string
	Criticality   string
	FilesChanged  int
	Migration     bool
	ConfigChange  bool
	CIStatus      string
	FailedTests   int
	Severity      string
	RiskScore     int
	RiskCategory  string
	HealthOverall string
	HealthAvail   bool
	Rollback      string
	Phase         string
}

func Evaluate(ctx context.Context, in Input, now time.Time) domain.PolicyEvaluation {
	version := readVersion()
	eval := domain.PolicyEvaluation{
		ReleaseID:     domain.ReleaseID(in.ReleaseID),
		PolicyVersion: version,
		Phase:         in.Phase,
		EvaluatedAt:   now.UTC(),
	}
	raw, err := json.Marshal(regoInput(in))
	if err != nil {
		return failClosed(eval, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return failClosed(eval, err)
	}
	module, err := policies.Bundle.ReadFile("release_gate.rego")
	if err != nil {
		return failClosed(eval, err)
	}
	r := rego.New(
		rego.Query("data.release_gate.rules"),
		rego.Module("release_gate.rego", string(module)),
		rego.Input(doc),
	)
	rs, err := r.Eval(ctx)
	if err != nil {
		return failClosed(eval, err)
	}
	fired := map[string]domain.PolicyRule{}
	if len(rs) > 0 && len(rs[0].Expressions) > 0 {
		rawRules, err := json.Marshal(rs[0].Expressions[0].Value)
		if err != nil {
			return failClosed(eval, err)
		}
		var items []map[string]any
		if err := json.Unmarshal(rawRules, &items); err != nil {
			return failClosed(eval, err)
		}
		for _, m := range items {
			rule := domain.PolicyRule{
				ID:           fmt.Sprint(m["id"]),
				Result:       fmt.Sprint(m["result"]),
				Message:      fmt.Sprint(m["message"]),
				InputExcerpt: fmt.Sprint(m["input_excerpt"]),
			}
			fired[rule.ID] = rule
		}
	}
	for _, id := range knownRules {
		if rule, ok := fired[id]; ok {
			eval.Rules = append(eval.Rules, rule)
			continue
		}
		skipped := strings.HasPrefix(id, "health.") && !in.HealthAvail
		if skipped {
			eval.Rules = append(eval.Rules, domain.PolicyRule{
				ID: id, Result: "PASS", Message: "skipped", InputExcerpt: "no_post_deploy_window", Skipped: true,
			})
			continue
		}
		eval.Rules = append(eval.Rules, domain.PolicyRule{ID: id, Result: "PASS", Message: "did not fire"})
	}
	eval.Result = aggregate(eval.Rules)
	return eval
}

func aggregate(rules []domain.PolicyRule) string {
	rank := map[string]int{"PASS": 0, "WARN": 1, "MANUAL_APPROVAL_REQUIRED": 2, "BLOCK": 3}
	best := "PASS"
	for _, r := range rules {
		if r.Skipped {
			continue
		}
		if rank[r.Result] > rank[best] {
			best = r.Result
		}
	}
	return best
}

func failClosed(eval domain.PolicyEvaluation, err error) domain.PolicyEvaluation {
	eval.Result = "BLOCK"
	eval.Rules = []domain.PolicyRule{{
		ID: engineRule, Result: "BLOCK", Message: "policy engine failed closed",
		InputExcerpt: "evaluation_error",
	}}
	_ = err
	return eval
}

func readVersion() string {
	b, err := policies.Bundle.ReadFile("VERSION")
	if err != nil {
		return "0.0.0+dev"
	}
	return strings.TrimSpace(string(b)) + "+dev"
}

func regoInput(in Input) map[string]any {
	return map[string]any{
		"release": map[string]any{"id": in.ReleaseID, "service_id": in.ServiceID},
		"service": map[string]any{"id": in.ServiceID, "criticality": in.Criticality},
		"change": map[string]any{
			"files_changed": in.FilesChanged, "migration_present": in.Migration, "config_change_present": in.ConfigChange,
		},
		"ci":       map[string]any{"failed_tests": in.FailedTests, "status": in.CIStatus},
		"security": map[string]any{"highest_severity": in.Severity},
		"risk":     map[string]any{"score": in.RiskScore, "category": in.RiskCategory},
		"health":   map[string]any{"overall": in.HealthOverall, "available": in.HealthAvail},
		"rollback": map[string]any{"status": in.Rollback},
	}
}
