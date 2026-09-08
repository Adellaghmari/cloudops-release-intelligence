package rollback

import "testing"

func TestReadyPartialUnknownNotReady(t *testing.T) {
	ready := Assess(Input{
		PreviousSuccessfulRelease: true, PreviousArtifact: true, PreviousImageDigest: true,
		ImageDigestApplicable: true, DeploymentTargetKnown: true, PreviousVersionRecorded: true,
	})
	if ready.Status != "READY" {
		t.Fatalf("%s", ready.Status)
	}
	partial := Assess(Input{
		PreviousSuccessfulRelease: true, PreviousArtifact: true, ImageDigestApplicable: true,
		PreviousImageDigest: true, DeploymentTargetKnown: true, PreviousVersionRecorded: true,
		MigrationPresent: true,
	})
	if partial.Status != "PARTIAL" {
		t.Fatalf("%s %+v", partial.Status, partial)
	}
	none := Assess(Input{})
	if none.Status != "NOT_READY" {
		t.Fatalf("%s", none.Status)
	}
	unknownMig := Assess(Input{
		PreviousSuccessfulRelease: true, PreviousArtifact: true, DeploymentTargetKnown: true,
		PreviousVersionRecorded: true, ImageDigestApplicable: false, MigrationPresent: true,
	})
	if unknownMig.Status != "PARTIAL" {
		t.Fatalf("unknown migration reversibility should be PARTIAL, got %s", unknownMig.Status)
	}
	noPrev := Assess(Input{PreviousArtifact: true, PreviousVersionRecorded: true, DeploymentTargetKnown: true})
	if noPrev.Status != "NOT_READY" {
		t.Fatalf("no previous release: %s", noPrev.Status)
	}
	missingArt := Assess(Input{PreviousSuccessfulRelease: true, PreviousVersionRecorded: true, DeploymentTargetKnown: true})
	if missingArt.Status != "NOT_READY" {
		t.Fatalf("missing artifact: %s", missingArt.Status)
	}
	contradict := Assess(Input{
		PreviousSuccessfulRelease: true, PreviousArtifact: false, ImageDigestApplicable: true,
		PreviousImageDigest: true, DeploymentTargetKnown: true, PreviousVersionRecorded: true,
	})
	if contradict.Status != "NOT_READY" {
		t.Fatalf("contradictory metadata: %s", contradict.Status)
	}
}
