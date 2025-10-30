package utils

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	pxray "go.opentelemetry.io/contrib/propagators/aws/xray"
	sxray "go.opentelemetry.io/contrib/samplers/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
)

func SetupTracing() error {
	ctx := context.Background()

	exporterEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if exporterEndpoint == "" {
		exporterEndpoint = "localhost:4317"
	}

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(exporterEndpoint))
	if err != nil {
		return fmt.Errorf("failed to create OTLP trace exporter: %v", err)
	}

	remoteSampler, err := sxray.NewRemoteSampler(ctx, "my-service-name", "ec2")
	if err != nil {
		return fmt.Errorf("failed to create X-Ray Remote Sampler: %v", err)
	}

	ec2Resource, err := ec2.NewResourceDetector().Detect(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect EC2 resource: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(remoteSampler),
		trace.WithBatcher(traceExporter),
		trace.WithResource(ec2Resource),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(pxray.Propagator{})

	return nil
}
