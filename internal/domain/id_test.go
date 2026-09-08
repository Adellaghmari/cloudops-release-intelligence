package domain

import "testing"

func TestParseIDs(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "service kebab", raw: "payments-service"},
		{name: "release prefixed", raw: "rel_01jabc"},
		{name: "event ulid-like", raw: "01ARZ3NDEKTSV4RRFFQ69G5FAV"},
		{name: "colon producer id", raw: "github:workflow:12345"},
		{name: "empty", raw: "", wantErr: true},
		{name: "spaces", raw: "bad id", wantErr: true},
		{name: "too long", raw: string(make([]byte, 129)), wantErr: true},
		{name: "slash", raw: "svc/payments", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := tt.raw
			if tt.name == "too long" {
				b := make([]byte, 129)
				for i := range b {
					b[i] = 'a'
				}
				raw = string(b)
			}
			_, err := ParseServiceID(raw)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil && !IsInvalid(err) {
				t.Fatalf("expected invalid error, got %v", err)
			}
		})
	}
}

func TestCommitSHA(t *testing.T) {
	if _, err := ParseCommitSHA("abc1234"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseCommitSHA("ZZZ"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := ParseCommitSHA("abcd"); err == nil {
		t.Fatal("expected short sha error")
	}
}

func TestGeneratedIDsValidate(t *testing.T) {
	ids := []interface{ Validate() error }{
		NewReleaseID(),
		NewDeploymentID(),
		NewEventID(),
		NewIncidentID(),
		NewCIRunID(),
		NewHealthSnapshotID(),
		NewDecisionID(),
	}
	for _, id := range ids {
		if err := id.Validate(); err != nil {
			t.Fatalf("%T: %v", id, err)
		}
	}
}

func TestTypedIDsAreDistinct(t *testing.T) {
	// Compile-time / usage check: converting between ID types must be explicit.
	var release ReleaseID = "rel_a"
	service := ServiceID(release)
	if service.String() != release.String() {
		t.Fatal("explicit conversion should preserve bytes")
	}
	if _, err := ParseEventID(release.String()); err != nil {
		t.Fatal(err)
	}
}
