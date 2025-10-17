//go:build connection

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.opentelemetry.io/contrib/detectors/aws/ec2"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	pxray "go.opentelemetry.io/contrib/propagators/aws/xray"
	sxray "go.opentelemetry.io/contrib/samplers/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
)

func init() {
	err := setupTracing()
	if err != nil {
		log.Fatalf("failed to setup tracing: %v", err)
	}
}

func setupTracing() error {
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

func response(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Headers":     "Content-Type",
			"Access-Control-Allow-Methods":     "OPTIONS, POST, GET, PUT, DELETE",
			"Access-Control-Allow-Credentials": "true",
		},
	}
	resp.StatusCode = status

	// Convert body to json data
	sBody, _ := json.Marshal(body)
	resp.Body = string(sBody)

	return &resp, nil
}

func handleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Parse the input event
	tableName := os.Getenv("TABLE_TEST_NAME")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS-SDK config, %v", err)
	}

	otelaws.AppendMiddlewares(&cfg.APIOptions)
	ddb := dynamodb.NewFromConfig(cfg)

	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}

	result, err := ddb.DescribeTable(ctx, input)
	if err != nil {
		log.Fatalf("unable to load secret, %s", err.Error())
	}

	log.Printf("Successfully loaded table %v", result.Table)

	// Return the region from environment variable
	return response(http.StatusOK,
		map[string]interface{}{
			"Region": os.Getenv("AWS_REGION"),
		})
}
