package seed

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
	log "github.com/sirupsen/logrus"
)

func MigratePermissions(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	client := db.GetClient()

	roleTable := os.Getenv("ROLE_DEFINITIONS_TABLE_NAME")
	if roleTable == "" {
		roleTable = "dev-role-definitions"
	}
	ownershipTable := os.Getenv("OWNERSHIP_POLICIES_TABLE_NAME")
	if ownershipTable == "" {
		ownershipTable = "dev-ownership-policies"
	}

	rdItems, err := scanTable(ctx, client, roleTable)
	if err != nil {
		return utils.ErrorResponse(500, "scan roles failed", err)
	}
	var roleDefs []*models.RoleDefinition
	if err := attributevalue.UnmarshalListOfMaps(rdItems, &roleDefs); err != nil {
		return utils.ErrorResponse(500, "unmarshal roles failed", err)
	}

	opItems, err := scanTable(ctx, client, ownershipTable)
	if err != nil {
		return utils.ErrorResponse(500, "scan ownerships failed", err)
	}
	var ownerships []*models.OwnershipPolicy
	if err := attributevalue.UnmarshalListOfMaps(opItems, &ownerships); err != nil {
		return utils.ErrorResponse(500, "unmarshal ownerships failed", err)
	}

	defsMap := make(map[string]models.ResourceDefinition)
	for _, d := range models.GetDefinitions() {
		defsMap[d.Id] = d
	}

	changedRoles := 0
	for _, rd := range roleDefs {
		newPerms := normalizePermList(rd.Permissions, defsMap)
		if !equalStringSlices(newPerms, rd.Permissions) {
			if _, err := db.UpdateRoleDefinitionPermissions(ctx, rd.Id, newPerms); err != nil {
				log.WithError(err).WithField("id", rd.Id).Warn("update role failed")
			} else {
				changedRoles++
			}
		}
	}

	changedOwnership := 0
	for _, op := range ownerships {
		newOwner := normalizePermList(op.OwnerPermissions, defsMap)
		newParent := normalizePermList(op.ParentOwnerPermissions, defsMap)
		if !equalStringSlices(newOwner, op.OwnerPermissions) || !equalStringSlices(newParent, op.ParentOwnerPermissions) {
			if _, err := db.UpsertOwnershipPolicy(ctx, op.TenantId, op.ResourceType, newOwner, newParent); err != nil {
				log.WithError(err).WithField("res", op.ResourceType).Warn("update ownership failed")
			} else {
				changedOwnership++
			}
		}
	}

	log.WithFields(log.Fields{"changedRoles": changedRoles, "changedOwnership": changedOwnership}).Info("Permissions migration done")
	return utils.SuccessResponse(200, "Permissions migration done", map[string]int{"roles": changedRoles, "ownership": changedOwnership})
}

func scanTable(ctx context.Context, client *dynamodb.Client, table string) ([]map[string]types.AttributeValue, error) {
	out := []map[string]types.AttributeValue{}
	var lastEvaluated map[string]types.AttributeValue
	for {
		in := &dynamodb.ScanInput{TableName: &table}
		if lastEvaluated != nil {
			in.ExclusiveStartKey = lastEvaluated
		}
		res, err := client.Scan(ctx, in)
		if err != nil {
			return nil, err
		}
		out = append(out, res.Items...)
		if res.LastEvaluatedKey == nil || len(res.LastEvaluatedKey) == 0 {
			break
		}
		lastEvaluated = res.LastEvaluatedKey
	}
	return out, nil
}

func normalizePermList(perms []string, defsMap map[string]models.ResourceDefinition) []string {
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		if p == "" {
			continue
		}
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 {
			out = append(out, p)
			continue
		}
		res, act := parts[0], strings.ToLower(parts[1])
		if _, ok := defsMap[res]; ok {
			out = append(out, models.GetPermission(res, act))
			continue
		}
		candidate := toSnake(res)
		if _, ok := defsMap[candidate]; ok {
			out = append(out, models.GetPermission(candidate, act))
			continue
		}
		lower := strings.ToLower(res)
		if _, ok := defsMap[lower]; ok {
			out = append(out, models.GetPermission(lower, act))
			continue
		}
		out = append(out, fmt.Sprintf("%s:%s", res, act))
	}
	return out
}

func toSnake(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := s[i-1]
			if prev != '_' {
				b.WriteRune('_')
			}
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
