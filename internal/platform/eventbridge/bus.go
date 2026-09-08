package eventbridge

import (
	"context"
	"encoding/json"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	awseb "github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

type API interface {
	PutEvents(ctx context.Context, params *awseb.PutEventsInput, optFns ...func(*awseb.Options)) (*awseb.PutEventsOutput, error)
}

type Bus struct {
	api    API
	bus    string
	source string
}

func New(api API, busName, source string) *Bus {
	return &Bus{api: api, bus: busName, source: source}
}

func (b *Bus) Publish(ctx context.Context, e events.Envelope) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = b.api.PutEvents(ctx, &awseb.PutEventsInput{
		Entries: []ebtypes.PutEventsRequestEntry{{
			EventBusName: aws.String(b.bus),
			Source:       aws.String(b.source),
			DetailType:   aws.String(string(e.EventType)),
			Detail:       aws.String(string(body)),
		}},
	})
	return err
}
