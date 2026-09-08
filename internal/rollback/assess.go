package rollback

import "github.com/Adellaghmari/cloudops-release-intelligence/internal/domain"

const ModelVersion = "rollback-v1"

type Signal struct {
	ID      string
	OK      bool
	NA      bool
	Detail  string
	Missing bool
}

type Assessment struct {
	Status       domain.RollbackStatus
	Signals      []Signal
	Missing      []string
	ModelVersion string
}

type Input struct {
	PreviousSuccessfulRelease bool
	PreviousArtifact          bool
	PreviousImageDigest       bool
	ImageDigestApplicable     bool
	DeploymentTargetKnown     bool
	PreviousVersionRecorded   bool
	MigrationPresent          bool
	MigrationReversible       *bool
	ConfigRollbackAvailable   bool
}

func Assess(in Input) Assessment {
	sigs := []Signal{
		flag("previous_successful_release", in.PreviousSuccessfulRelease, true, "prior successful release recorded"),
		flag("previous_artifact", in.PreviousArtifact, true, "prior artifact URI recorded"),
		digest(in),
		flag("deployment_target_known", in.DeploymentTargetKnown, true, "deployment target known"),
		flag("previous_version_recorded", in.PreviousVersionRecorded, true, "immutable prior version recorded"),
		migration(in),
		{ID: "config_rollback", OK: in.ConfigRollbackAvailable, Detail: "optional prior config hash"},
	}
	missing := []string{}
	requiredMissing := 0
	unknown := 0
	hard := false
	for _, s := range sigs {
		if s.NA || s.ID == "config_rollback" {
			continue
		}
		if s.Missing {
			unknown++
		}
		if !s.OK && !s.Missing {
			missing = append(missing, s.ID)
			requiredMissing++
			if s.ID == "previous_artifact" || s.ID == "previous_successful_release" {
				hard = true
			}
		}
	}
	passed := 0
	for _, s := range sigs {
		if s.NA || s.ID == "config_rollback" {
			continue
		}
		if s.OK && !s.Missing {
			passed++
		}
	}
	status := domain.RollbackReady
	switch {
	case hard:
		status = domain.RollbackNotReady
	case requiredMissing > 0:
		status = domain.RollbackPartial
	case unknown > 0 && passed == 0:
		status = domain.RollbackUnknown
	case unknown > 0:
		status = domain.RollbackPartial
	default:
		status = domain.RollbackReady
	}
	return Assessment{Status: status, Signals: sigs, Missing: missing, ModelVersion: ModelVersion}
}

func flag(id string, ok, _ bool, detail string) Signal {
	return Signal{ID: id, OK: ok, Detail: detail, Missing: false}
}

func digest(in Input) Signal {
	if !in.ImageDigestApplicable {
		return Signal{ID: "previous_image_digest", OK: true, NA: true, Detail: "not applicable for non-container artifact"}
	}
	return Signal{ID: "previous_image_digest", OK: in.PreviousImageDigest, Detail: "previous image digest"}
}

func migration(in Input) Signal {
	if !in.MigrationPresent {
		return Signal{ID: "migration_reversibility", OK: true, NA: true, Detail: "no migration in this change"}
	}
	if in.MigrationReversible == nil {
		return Signal{ID: "migration_reversibility", OK: false, Missing: true, Detail: "migration present; reversibility unknown"}
	}
	return Signal{ID: "migration_reversibility", OK: *in.MigrationReversible, Detail: "migration reversibility metadata"}
}
