package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceContextHandlerAddsTraceFields(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(&traceContextHandler{
		Handler: slog.NewJSONHandler(&output, nil),
	})
	provider := sdktrace.NewTracerProvider()
	defer func() { _ = provider.Shutdown(context.Background()) }()

	ctx, span := provider.Tracer("logger-test").Start(context.Background(), "request")
	logger.InfoContext(ctx, "handled request", "request_id", "request-123")
	span.End()

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if record["trace_id"] == "" {
		t.Fatal("log record does not contain trace_id")
	}
	if record["span_id"] == "" {
		t.Fatal("log record does not contain span_id")
	}
	if record["request_id"] != "request-123" {
		t.Fatalf("request_id = %v, want request-123", record["request_id"])
	}
}

func TestTraceContextHandlerOmitsInvalidTraceFields(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(&traceContextHandler{
		Handler: slog.NewJSONHandler(&output, nil),
	})

	logger.InfoContext(context.Background(), "application started")

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if _, exists := record["trace_id"]; exists {
		t.Fatal("log record contains trace_id outside a span")
	}
	if _, exists := record["span_id"]; exists {
		t.Fatal("log record contains span_id outside a span")
	}
}
