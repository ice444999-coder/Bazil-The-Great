/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package observability

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		for _, fn := range shutdownFuncs {
			if err := fn(ctx); err != nil {
				log.Printf("Error shutting down: %v", err)
			}
		}
		shutdownFuncs = nil
		return nil
	}

	handleErr := func(inErr error) {
		err = inErr
		if err != nil {
			shutdown(ctx)
		}
	}

	// Set up trace exporter
	traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		handleErr(err)
		return
	}

	// Set up trace provider
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("ares-api"),
		),
	)
	if err != nil {
		handleErr(err)
		return
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	return
}
