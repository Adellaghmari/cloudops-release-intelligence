package s3evidence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Adellaghmari/cloudops-release-intelligence/internal/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type Writer struct {
	api    API
	bucket string
}

func New(api API, bucket string) *Writer {
	return &Writer{api: api, bucket: bucket}
}

func (w *Writer) PutEvent(ctx context.Context, e events.Envelope) error {
	if w == nil || w.bucket == "" {
		return nil
	}
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	at := e.OccurredAt
	if at.IsZero() {
		at = time.Now().UTC()
	}
	key := fmt.Sprintf("events/%s/%s.json", at.UTC().Format("2006/01"), e.EventID)
	_, err = w.api.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(w.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String("application/json"),
	})
	return err
}
