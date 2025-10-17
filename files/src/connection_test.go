//go:build connection

package main

import (
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

func TestConnection(t *testing.T) {
	resp, err := handleRequest(t.Context(), events.APIGatewayProxyRequest{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
