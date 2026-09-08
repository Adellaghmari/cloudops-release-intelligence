package dynamo

import (
	"time"

	"github.com/adell/cloudops-release-intelligence/internal/domain"
)

// Key helpers stay inside the persistence adapter. Domain types never see PK/SK.

func servicePK(id domain.ServiceID) string { return "SERVICE#" + id.String() }
func releasePK(id domain.ReleaseID) string { return "RELEASE#" + id.String() }
func incidentPK(id domain.IncidentID) string {
	return "INCIDENT#" + id.String()
}
func idemPK(id domain.EventID) string { return "IDEM#" + id.String() }

func metaSK() string                         { return "META" }
func depSK(to domain.ServiceID) string       { return "DEP#" + to.String() }
func deploySK(id domain.DeploymentID) string { return "DEPLOY#" + id.String() }
func commitSK() string                       { return "COMMIT#META" }
func ciSK(id domain.CIRunID) string          { return "CI#" + id.String() }
func decisionSK(id domain.DecisionID) string {
	return "DECISION#" + id.String()
}
func eventSK(at time.Time, id domain.EventID) string {
	return "EVENT#" + formatTime(at) + "#" + id.String()
}
func healthSK(end time.Time, id domain.HealthSnapshotID) string {
	return "HEALTH#" + formatTime(end) + "#" + id.String()
}

func typeServiceGSI1() (pk, skPrefix string) { return "TYPE#SERVICE", "NAME#" }
func serviceNameGSI1(name string) (pk, sk string) {
	return "TYPE#SERVICE", "NAME#" + name
}

func releaseByServiceGSI1(serviceID domain.ServiceID, created time.Time, id domain.ReleaseID) (pk, sk string) {
	return "SERVICE#" + serviceID.String(), "RELEASE#" + formatTime(created) + "#" + id.String()
}

func releaseTypeGSI2(created time.Time, id domain.ReleaseID) (pk, sk string) {
	return "TYPE#RELEASE", "TIME#" + formatTime(created) + "#" + id.String()
}

func depByGSI1(to, from domain.ServiceID) (pk, sk string) {
	return "SERVICE#" + to.String(), "DEPBY#" + from.String()
}

func healthByReleaseGSI1(releaseID domain.ReleaseID, at time.Time) (pk, sk string) {
	return "RELEASE#" + releaseID.String(), "HEALTH#" + formatTime(at)
}

func incidentByServiceGSI1(serviceID domain.ServiceID, opened time.Time) (pk, sk string) {
	return "SERVICE#" + serviceID.String(), "INCIDENT#" + formatTime(opened)
}

func incidentByReleaseGSI2(releaseID domain.ReleaseID, opened time.Time) (pk, sk string) {
	return "RELEASE#" + releaseID.String(), "INCIDENT#" + formatTime(opened)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
