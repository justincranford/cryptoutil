// Copyright (c) 2025 Justin Cranford
//
//

package telemetry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	stdoutLogExporter "log/slog"

	grpcTraceExporterotlptracegrpc "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	httpTraceExporterotlptracehttp "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	attributeApi "go.opentelemetry.io/otel/attribute"
	propagationApi "go.opentelemetry.io/otel/propagation"
	traceApi "go.opentelemetry.io/otel/trace"

	traceSdk "go.opentelemetry.io/otel/sdk/trace"

	oltpSemanticConventions "go.opentelemetry.io/otel/semconv/v1.30.0"
)

// TelemetryService is a composite of OpenTelemetry providers for Logs, Metrics, and Traces.

func initTextMapPropagator() *propagationApi.TextMapPropagator {
	textMapPropagator := propagationApi.NewCompositeTextMapPropagator(
		propagationApi.TraceContext{},
		propagationApi.Baggage{},
	)

	return &textMapPropagator
}

func doExampleTracesSpans(ctx context.Context, tracesProvider *traceSdk.TracerProvider, slogger *stdoutLogExporter.Logger) {
	tracer := tracesProvider.Tracer("fiber-tracer")
	spanCtx := doExampleTraceSpan(ctx, tracer, slogger, "sample parent trace and span")

	doExampleTraceSpan(spanCtx, tracer, slogger, "sample child trace and span")
}

func doExampleTraceSpan(ctx context.Context, tracer traceApi.Tracer, slogger *stdoutLogExporter.Logger, message string) context.Context {
	spanCtx, span := tracer.Start(ctx, "test-span")
	slogger.Debug(message, "traceid", span.SpanContext().TraceID(), "traceid", span.SpanContext().SpanID())

	return spanCtx
}

func getOtelMetricsTracesAttributes(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) []attributeApi.KeyValue {
	return []attributeApi.KeyValue{
		oltpSemanticConventions.DeploymentID(settings.OTLPEnvironment),   // deployment.environment.name (e.g. local-standalone, adhoc, dev, qa, preprod, prod)
		oltpSemanticConventions.HostName(settings.OTLPHostname),          // service.instance.id (e.g. 12)
		oltpSemanticConventions.ServiceName(settings.OTLPService),        // service.name (e.g. cryptoutil)
		oltpSemanticConventions.ServiceVersion(settings.OTLPVersion),     // service.version (e.g. 0.0.1, 1.0.2, 2.1.0)
		oltpSemanticConventions.ServiceInstanceID(settings.OTLPInstance), // service.instance.id (e.g. 12, uuidV7)
	}
}

func getOtelLogsAttributes(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) []attributeApi.KeyValue {
	return getOtelMetricsTracesAttributes(settings) // same (for now)
}

func getSlogStdoutAttributes(settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) []stdoutLogExporter.Attr {
	otelAttrs := getOtelLogsAttributes(settings)
	slogAttrs := make([]stdoutLogExporter.Attr, 0, len(otelAttrs))

	for _, otelLogAttr := range otelAttrs {
		slogAttrs = append(slogAttrs, stdoutLogExporter.String(string(otelLogAttr.Key), otelLogAttr.Value.AsString()))
	}

	return slogAttrs
}

func parseProtocolAndEndpoint(otlpEndpoint *string) (bool, bool, bool, bool, *string, error) {
	if after, ok := strings.CutPrefix(*otlpEndpoint, "http://"); ok {
		return true, false, false, false, &after, nil
	} else if after, ok := strings.CutPrefix(*otlpEndpoint, "https://"); ok {
		return false, true, false, false, &after, nil
	} else if after, ok := strings.CutPrefix(*otlpEndpoint, "grpc://"); ok {
		return false, false, true, false, &after, nil
	} else if after, ok := strings.CutPrefix(*otlpEndpoint, "grpcs://"); ok {
		return false, false, false, true, &after, nil
	}

	return false, false, false, false, nil, fmt.Errorf("invalid OTLP endpoint protocol, must start with https://, grpcs://, http://, or grpc://")
}

// checkSidecarHealth performs a connectivity check to the OTLP sidecar during startup.
func checkSidecarHealth(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) error {
	// Parse the endpoint to determine protocol and address
	isHTTP, isHTTPS, isGRPC, isGRPCS, endpoint, err := parseProtocolAndEndpoint(&settings.OTLPEndpoint)
	if err != nil {
		return fmt.Errorf("failed to parse OTLP endpoint: %w", err)
	}

	// For now, we do a basic connectivity check by attempting to create an exporter
	// This will fail if the sidecar is not reachable
	if isGRPC {
		_, _ = grpcTraceExporterotlptracegrpc.New(ctx,
			grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint),
			grpcTraceExporterotlptracegrpc.WithInsecure())
	} else if isGRPCS {
		_, _ = grpcTraceExporterotlptracegrpc.New(ctx,
			grpcTraceExporterotlptracegrpc.WithEndpoint(*endpoint))
	} else if isHTTP {
		_, err = httpTraceExporterotlptracehttp.New(ctx,
			httpTraceExporterotlptracehttp.WithEndpoint(*endpoint),
			httpTraceExporterotlptracehttp.WithInsecure())
		if err != nil {
			return fmt.Errorf("HTTP sidecar connectivity check failed: %w", err)
		}
	} else if isHTTPS {
		_, _ = httpTraceExporterotlptracehttp.New(ctx,
			httpTraceExporterotlptracehttp.WithEndpoint(*endpoint))
	}

	return nil
}

// checkSidecarHealthWithRetry performs connectivity check to OTLP sidecar with retry logic, before init logsProvider; caller must log results.
func checkSidecarHealthWithRetry(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) ([]error, error) {
	maxRetries := cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries
	retryDelay := cryptoutilSharedMagic.DefaultSidecarHealthCheckRetryDelay

	var intermediateErrs []error

	for attempt := range maxRetries + 1 {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return intermediateErrs, fmt.Errorf("context cancelled during sidecar health check retry: %w", ctx.Err())
			case <-time.After(retryDelay):
				// Continue with retry
			}
		}

		err := checkSidecarHealth(ctx, settings)
		if err == nil {
			return intermediateErrs, nil
		}

		intermediateErrs = append(intermediateErrs, err)
	}

	return intermediateErrs, fmt.Errorf("sidecar health check failed after %d attempts (%d failures): %w", maxRetries+1, len(intermediateErrs), errors.Join(intermediateErrs...))
}

// CheckSidecarHealth performs a connectivity check to the OTLP sidecar.
func (s *TelemetryService) CheckSidecarHealth(ctx context.Context) error {
	if s.settings.OTLPEnabled {
		err := checkSidecarHealth(ctx, s.settings)
		if err != nil {
			return fmt.Errorf("sidecar health check failed: %w", err)
		}
	}

	return nil
}
