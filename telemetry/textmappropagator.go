package telemetry

import "go.opentelemetry.io/otel/propagation"

func InitTextMapPropagator() *propagation.TextMapPropagator {
	textMapPropagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	return &textMapPropagator
}
