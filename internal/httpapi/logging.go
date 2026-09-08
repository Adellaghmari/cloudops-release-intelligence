package httpapi

import (
	"context"
	"log/slog"
)

type ctxKey string

const (
	ctxRequestID     ctxKey = "request_id"
	ctxCorrelationID ctxKey = "correlation_id"
)

func WithIDs(ctx context.Context, requestID, correlationID string) context.Context {
	ctx = context.WithValue(ctx, ctxRequestID, requestID)
	ctx = context.WithValue(ctx, ctxCorrelationID, correlationID)
	return ctx
}

func RequestIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxRequestID).(string)
	return v
}

func CorrelationIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxCorrelationID).(string)
	return v
}

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := RequestIDFrom(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	if id := CorrelationIDFrom(ctx); id != "" {
		r.AddAttrs(slog.String("correlation_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}

func NewLogger(level slog.Level) *slog.Logger {
	h := slog.NewJSONHandler(osStdout(), &slog.HandlerOptions{Level: level})
	return slog.New(contextHandler{Handler: h})
}
