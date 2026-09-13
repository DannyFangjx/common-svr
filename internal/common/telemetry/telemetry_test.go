package telemetry

import (
	"context"
	"testing"

	"common-svr/internal/common/config"

	"go.opentelemetry.io/otel/trace"
)

func TestNewWithoutEndpointCreatesLocalTrace(t *testing.T) {
	provider, err := New(context.Background(), config.Telemetry{
		Protocol:    "grpc",
		SampleRatio: 1,
	}, "common-svr-test", "test")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer func() { _ = provider.Shutdown(context.Background()) }()

	ctx, span := provider.Tracer("telemetry-test").Start(context.Background(), "request")
	defer span.End()

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		t.Fatal("span context is invalid when exporter endpoint is empty")
	}
}
