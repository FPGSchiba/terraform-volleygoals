package resource_definitions

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/aws/aws-lambda-go/events"
    "github.com/fpgschiba/volleygoals/models"
)

func TestGetResourceDefinitions(t *testing.T) {
    resp, err := GetResourceDefinitions(context.Background(), events.APIGatewayProxyRequest{})
    if err != nil {
        t.Fatalf("handler returned error: %v", err)
    }
    if resp.StatusCode != 200 {
        t.Fatalf("expected status 200, got %d", resp.StatusCode)
    }
    var defs []models.ResourceDefinition
    if err := json.Unmarshal([]byte(resp.Body), &defs); err != nil {
        t.Fatalf("invalid body json: %v", err)
    }
    if len(defs) == 0 {
        t.Fatalf("expected at least one resource definition")
    }
    for _, d := range defs {
        if d.Id == "" {
            t.Fatalf("definition missing id: %+v", d)
        }
        if d.Name == "" {
            t.Fatalf("definition missing name: %+v", d)
        }
        if len(d.Actions) == 0 {
            t.Fatalf("definition actions empty for %s", d.Id)
        }
    }
}

