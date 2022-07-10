package common

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tracerProvider *sdktrace.TracerProvider

func InitOtelProvider() {
	Debug("Initializing Opentelemetry Provider")

	ctx := context.Background()

	// Configure a new exporter using environment variables for sending data to Honeycomb over gRPC.
	exporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		Error("failed to initialize exporter: %v", err)
	}

	// Create a new tracer provider with a batch span processor and the otlp exporter.
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	// Set the Tracer Provider and the W3C Trace Context propagator as globals
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Register the trace context and baggage propagators so data is propagated across services/processes.
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	Debug("Opentelemetry Provider Initialized")
}

func CloseOtelProvider() {
	Debug("Shutting down Opentelemetry Provider")
	// Shutdown will flush any remaining spans and shut down the exporter.
	err := tracerProvider.Shutdown(context.Background())

	if err != nil {
		Error("Failed to shutdown TracerProvider: %v", err)
	}

}
