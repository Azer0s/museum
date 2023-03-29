package observability

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"museum/config"
)

func NewSpanExporter(config config.Config) tracesdk.SpanExporter {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://" + config.GetJaegerHost() + "/api/traces")))
	if err != nil {
		panic(err)
	}
	return exp
}

func NewTracerProvider(exporter tracesdk.SpanExporter, config config.Config) trace.TracerProvider {
	return tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("museum"),
		)),
	)
}
