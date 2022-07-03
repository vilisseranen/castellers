package common

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var tracerProvider *sdktrace.TracerProvider

func InitOtelProvider() {
	Debug("Initializing Opentelemetry Provider")

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("castellers"),
		),
	)
	if err != nil {
		Error("Failed to create resource: %v", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:4317", GetConfigString("otlp_endpoint"))),
		otlptracegrpc.WithDialOption(),
	)
	if err != nil {
		Error("Failed to create trace exporter: %v", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

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
