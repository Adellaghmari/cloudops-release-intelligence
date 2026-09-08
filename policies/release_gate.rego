package release_gate

import rego.v1

# Fired rules only. Go aggregates precedence and fills PASS/skipped.
rules contains result if {
	input.ci.failed_tests > 0
	result := {
		"id": "tests.must_pass",
		"result": "BLOCK",
		"message": "Attached CI run has failing tests",
		"input_excerpt": sprintf("failed_tests=%v status=%v", [input.ci.failed_tests, input.ci.status]),
	}
}

rules contains result if {
	input.security.highest_severity == "CRITICAL"
	result := {
		"id": "security.critical_vuln",
		"result": "BLOCK",
		"message": "Critical security finding is present",
		"input_excerpt": sprintf("highest_severity=%v", [input.security.highest_severity]),
	}
}

rules contains result if {
	input.security.highest_severity == "HIGH"
	result := {
		"id": "security.high_vuln",
		"result": "WARN",
		"message": "High severity security finding is present",
		"input_excerpt": sprintf("highest_severity=%v", [input.security.highest_severity]),
	}
}

rules contains result if {
	input.risk.score >= 50
	input.risk.score < 75
	result := {
		"id": "risk.high_threshold",
		"result": "WARN",
		"message": "Release risk score is HIGH",
		"input_excerpt": sprintf("score=%v category=%v", [input.risk.score, input.risk.category]),
	}
}

rules contains result if {
	input.risk.score >= 75
	result := {
		"id": "risk.critical_threshold",
		"result": "MANUAL_APPROVAL_REQUIRED",
		"message": "Release risk score is CRITICAL",
		"input_excerpt": sprintf("score=%v category=%v", [input.risk.score, input.risk.category]),
	}
}

rules contains result if {
	input.change.migration_present
	input.rollback.status != "READY"
	result := {
		"id": "migration.rollback_plan",
		"result": "WARN",
		"message": "Database migration without READY rollback metadata",
		"input_excerpt": sprintf("migration_present=%v rollback=%v", [input.change.migration_present, input.rollback.status]),
	}
}

rules contains result if {
	input.service.criticality == "CRITICAL"
	result := {
		"id": "critical_service.change",
		"result": "MANUAL_APPROVAL_REQUIRED",
		"message": "Critical service change requires manual approval",
		"input_excerpt": sprintf("service=%v criticality=%v", [input.service.id, input.service.criticality]),
	}
}

rules contains result if {
	input.health.available
	input.health.overall == "DEGRADED"
	result := {
		"id": "health.availability_slo",
		"result": "BLOCK",
		"message": "Post-deploy health is DEGRADED",
		"input_excerpt": sprintf("overall=%v", [input.health.overall]),
	}
}

rules contains result if {
	input.health.available
	input.health.overall == "SEVERELY_DEGRADED"
	result := {
		"id": "health.severe_regression",
		"result": "BLOCK",
		"message": "Post-deploy health is SEVERELY_DEGRADED",
		"input_excerpt": sprintf("overall=%v", [input.health.overall]),
	}
}
