package domain

import "testing"

func TestParseEnums(t *testing.T) {
	tests := []struct {
		name    string
		parse   func() error
		wantErr bool
	}{
		{name: "source live", parse: func() error { _, err := ParseDataSource("live"); return err }},
		{name: "source synthetic", parse: func() error { _, err := ParseDataSource("synthetic"); return err }},
		{name: "source production rejected", parse: func() error { _, err := ParseDataSource("production"); return err }, wantErr: true},
		{name: "criticality lower", parse: func() error { _, err := ParseCriticality("critical"); return err }},
		{name: "criticality bad", parse: func() error { _, err := ParseCriticality("urgent"); return err }, wantErr: true},
		{name: "environment prod", parse: func() error { _, err := ParseEnvironment("prod"); return err }},
		{name: "environment production rejected", parse: func() error { _, err := ParseEnvironment("production"); return err }, wantErr: true},
		{name: "release status", parse: func() error { _, err := ParseReleaseStatus("deployed"); return err }},
		{name: "release status bad", parse: func() error { _, err := ParseReleaseStatus("SHIPPED"); return err }, wantErr: true},
		{name: "deployment status", parse: func() error { _, err := ParseDeploymentStatus("succeeded"); return err }},
		{name: "ci status", parse: func() error { _, err := ParseCIRunStatus("failed"); return err }},
		{name: "incident open", parse: func() error { _, err := ParseIncidentStatus("open"); return err }},
		{name: "dependency async", parse: func() error { _, err := ParseDependencyKind("async"); return err }},
		{name: "health window", parse: func() error { _, err := ParseHealthWindowKind("baseline"); return err }},
		{name: "decision", parse: func() error { _, err := ParseDecisionKind("manual_approve"); return err }},
		{name: "producer", parse: func() error { _, err := ParseEventProducer("github-actions"); return err }},
		{name: "event type", parse: func() error { _, err := ParseEventType("deployment.succeeded"); return err }},
		{name: "event type unknown", parse: func() error { _, err := ParseEventType("magic.happened"); return err }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parse()
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}
			if err != nil && !IsInvalid(err) {
				t.Fatalf("expected invalid, got %v", err)
			}
		})
	}
}

func TestDataSourceValues(t *testing.T) {
	if DataSourceLive == DataSourceSynthetic {
		t.Fatal("live and synthetic must remain distinct")
	}
}
