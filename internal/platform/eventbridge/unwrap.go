package eventbridge

import (
	"encoding/json"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
)

type wrappedEvent struct {
	Detail json.RawMessage `json:"detail"`
}

func UnwrapSQSBody(body string) (events.Envelope, error) {
	var env events.Envelope
	if err := json.Unmarshal([]byte(body), &env); err == nil && env.EventID != "" {
		return env, nil
	}
	var wrap wrappedEvent
	if err := json.Unmarshal([]byte(body), &wrap); err != nil {
		return events.Envelope{}, err
	}
	if err := json.Unmarshal(wrap.Detail, &env); err != nil {
		return events.Envelope{}, err
	}
	return env, nil
}
