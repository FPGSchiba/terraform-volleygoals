//go:build getTeams

package router

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestConnection(t *testing.T) {
	resp, err := HandleRequest(t.Context(), events.APIGatewayProxyRequest{})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}
