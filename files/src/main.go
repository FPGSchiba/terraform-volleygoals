//go:build !local

package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/mail"
	"github.com/fpgschiba/volleygoals/router"
	"github.com/fpgschiba/volleygoals/storage"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

var tp *trace.TracerProvider

func init() {
	// Ensure logger is configured as early as possible
	utils.InitLogger()

	ctx := context.Background()

	// Setup AWS SDK
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config: %v", err)
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)

	db.InitClient(&cfg)
	mail.InitClient(&cfg)
	storage.InitClient(&cfg)
	users.InitClient(&cfg)

	// Initialize OpenTelemetry tracing (sets global tracer provider)
	SetupTracing(ctx)
}

func SetupTracing(ctx context.Context) {
	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("volleygoals"),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// Use HTTP exporter instead of gRPC (more reliable in Lambda)
	exporter, err := otlptrace.New(ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("localhost:4318"),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatalf("failed to create OTLP exporter: %v", err)
	}

	// Create TracerProvider with aggressive batching settings for Lambda
	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(100*time.Millisecond), // Flush quickly
			trace.WithMaxExportBatchSize(10),             // Small batches
			trace.WithMaxQueueSize(100),                  // Small queue
		),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithIDGenerator(xray.NewIDGenerator()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(xray.Propagator{})
}

func main() {
	// Create base handler that delegates to router.HandleRequest
	baseHandler := func(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		// Handle the request
		response, err := router.HandleRequest(ctx, event)

		// Force flush with timeout before returning (guard tp against nil)
		if tp != nil {
			flushCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if flushErr := tp.ForceFlush(flushCtx); flushErr != nil {
				log.WithError(flushErr).Warn("Failed to flush spans")
			}
		}

		return response, err
	}

	// Wrap handler with JSONLogLambdaWrapper for consistent logging in Lambda
	h := utils.JSONLogLambdaWrapper(baseHandler)

	lambda.Start(h)
}
