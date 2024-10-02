package observability

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"museum/config"
)

func NewSpanExporter(config config.Config, log *zap.SugaredLogger) tracesdk.SpanExporter {
	if config.GetJaegerHost() == "" {
		log.Warn("jaeger host not set, using noop exporter")
		return &tracetest.NoopExporter{}
	}

	client, err := grpc.NewClient(config.GetJaegerHost(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Panicw("failed to create gRPC connection to collector", "error", err)
	}

	exp, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(client))
	if err != nil {
		log.Panicw("failed to create jaeger exporter", "error", err)
	}
	return exp
}

type TracerProviderFactory struct {
	Build func(serviceName string) trace.TracerProvider
}

func NewTracerProviderFactory(exporter tracesdk.SpanExporter, config config.Config) *TracerProviderFactory {
	return &TracerProviderFactory{
		Build: func(serviceName string) trace.TracerProvider {
			return tracesdk.NewTracerProvider(
				// Always be sure to batch in production.
				tracesdk.WithBatcher(exporter),
				// Record information about this application in a Resource.
				tracesdk.WithResource(resource.NewWithAttributes(
					semconv.SchemaURL,
					semconv.ServiceName(serviceName),
					attribute.String("environment", config.GetEnvironment()),
				)),
			)
		},
	}
}

func NewDefaultTracerProvider(factory *TracerProviderFactory) trace.TracerProvider {
	return factory.Build("museum")
}
