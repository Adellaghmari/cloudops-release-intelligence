package eventbridge

import "testing"

func TestUnwrapDirectAndWrapped(t *testing.T) {
	direct := `{"event_id":"evt_1","event_type":"ci.run.completed","schema_version":"1.0"}`
	env, err := UnwrapSQSBody(direct)
	if err != nil || env.EventID != "evt_1" {
		t.Fatalf("direct=%v err=%v", env, err)
	}
	wrapped := `{"detail":{"event_id":"evt_2","event_type":"deployment.succeeded","schema_version":"1.0"}}`
	env, err = UnwrapSQSBody(wrapped)
	if err != nil || env.EventID != "evt_2" {
		t.Fatalf("wrapped=%v err=%v", env, err)
	}
}
