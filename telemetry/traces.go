package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/sdk/resource"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

func InitTraces(ctx context.Context, enableOtel bool, enableStdout bool) *trace.TracerProvider {
	var tracesOptions []trace.TracerProviderOption

	otelMeterTracerResource, err := resource.New(ctx, resource.WithAttributes(otelMetricsTracesAttributes...))
	ifErrorLogAndExit("create Otel GRPC traces resource failed: %v", err)
	tracesOptions = append(tracesOptions, trace.WithResource(otelMeterTracerResource))

	if enableOtel {
		tracerOtelGrpc, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(OtelGrpcPush), otlptracegrpc.WithInsecure())
		ifErrorLogAndExit("create Otel GRPC traces failed: %v", err)
		tracesOptions = append(tracesOptions, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(tracerOtelGrpc, trace.WithBatchTimeout(500*time.Millisecond))))
	}

	if enableStdout {
		stdoutTraces, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		ifErrorLogAndExit("create STDOUT traces failed: %v", err)
		tracesOptions = append(tracesOptions, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(stdoutTraces, trace.WithBatchTimeout(500*time.Millisecond))))
	}

	return trace.NewTracerProvider(tracesOptions...)
}

func DoTraceExample(ctx context.Context, telemetryService *Service) {
	exampleTrace := telemetryService.TracesProvider.Tracer("example-trace")
	telemetryService.Slogger.Info("exampleTrace", "trace", exampleTrace)

	// simulate time spent in parent function, before calling child 1 function
	exampleParentSpanContext, exampleParentSpan := exampleTrace.Start(ctx, "example-parent-span")
	time.Sleep(5 * time.Millisecond)
	exampleParentSpan.End()
	telemetryService.Slogger.Info("exampleParentSpan", "ctx", exampleParentSpanContext, "span", exampleParentSpan)

	// simulate time spent in child 1 function
	exampleChildSpanContext1, exampleChildSpan1 := exampleTrace.Start(exampleParentSpanContext, "example-child-span-1")
	time.Sleep(10 * time.Millisecond)
	defer exampleChildSpan1.End()
	telemetryService.Slogger.Info("exampleChildSpan1", "ctx", exampleChildSpanContext1, "span", exampleChildSpan1)

	// simulate time spent in parent function, before calling child 2 function
	time.Sleep(5 * time.Millisecond)

	// simulate time spent in child 2 function
	exampleChildSpanContext2, exampleChildSpan2 := exampleTrace.Start(exampleParentSpanContext, "example-child-span-2")
	time.Sleep(15 * time.Millisecond)
	defer exampleChildSpan2.End()
	telemetryService.Slogger.Info("exampleChildSpan2", "ctx", exampleChildSpanContext2, "span", exampleChildSpan2)

	// simulate time spent in parent function, before returning
	time.Sleep(5 * time.Millisecond)
}
