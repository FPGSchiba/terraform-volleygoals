//go:build ignore

// migrate_permissions scans RoleDefinition and OwnershipPolicy records and
// normalizes permission strings to the canonical `<resource>:<action>` format
// driven by `models.GetDefinitions()`.
//
// Usage (local dev tables):
//   go run -tags local files/src/scripts/migrate_permissions/main.go -apply=false
// Flags:
//   -apply (bool) : when true, writes updates to DB. Default false (dry-run).
//   -role-table (string): override role definitions table name (default dev-role-definitions)
//   -ownership-table (string): override ownership policies table name (default dev-ownership-policies)
//   -backup (string): optional path to write a JSON backup of scanned records
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "os"
    "strings"

    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "github.com/fpgschiba/volleygoals/db"
    "github.com/fpgschiba/volleygoals/models"
)

func main() {
    apply := flag.Bool("apply", false, "When true, apply changes to DB. Default false (dry-run).")
    roleTable := flag.String("role-table", "dev-role-definitions", "DynamoDB table name for role definitions")
    ownershipTable := flag.String("ownership-table", "dev-ownership-policies", "DynamoDB table name for ownership policies")
    backupPath := flag.String("backup", "", "Optional path to write JSON backup of scanned records")
    flag.Parse()

    ctx := context.Background()
    db.InitClient(nil)
    client := db.GetClient()

    // Scan role definitions
    fmt.Println("Scanning role definitions table:", *roleTable)
    rdItems, err := scanTable(ctx, client, *roleTable)
    if err != nil {
        fmt.Fprintf(os.Stderr, "scan role definitions: %v\n", err)
        os.Exit(1)
    }
    var roleDefs []*models.RoleDefinition
    if err := attributevalue.UnmarshalListOfMaps(rdItems, &roleDefs); err != nil {
        fmt.Fprintf(os.Stderr, "unmarshal role defs: %v\n", err)
        os.Exit(1)
    }

    // Scan ownership policies
    fmt.Println("Scanning ownership policies table:", *ownershipTable)
    opItems, err := scanTable(ctx, client, *ownershipTable)
    if err != nil {
        fmt.Fprintf(os.Stderr, "scan ownership policies: %v\n", err)
        os.Exit(1)
    }
    var ownerships []*models.OwnershipPolicy
    if err := attributevalue.UnmarshalListOfMaps(opItems, &ownerships); err != nil {
        fmt.Fprintf(os.Stderr, "unmarshal ownership policies: %v\n", err)
        os.Exit(1)
    }

    // Optional backup
    if *backupPath != "" {
        backup := map[string]interface{}{"role_definitions": roleDefs, "ownership_policies": ownerships}
        b, _ := json.MarshalIndent(backup, "", "  ")
        if err := os.WriteFile(*backupPath, b, 0644); err != nil {
            fmt.Fprintf(os.Stderr, "write backup: %v\n", err)
            // continue even if backup fails
        } else {
            fmt.Println("Wrote backup to", *backupPath)
        }
    }

    // Normalize permissions for role definitions
    defsMap := make(map[string]models.ResourceDefinition)
    for _, d := range models.GetDefinitions() {
        defsMap[d.Id] = d
    }

    changedRoles := 0
    for _, rd := range roleDefs {
        newPerms := normalizePermList(rd.Permissions, defsMap)
        if !equalStringSlices(newPerms, rd.Permissions) {
            fmt.Printf("Role %s (%s) permissions will be updated: %d -> %d\n", rd.Name, rd.Id, len(rd.Permissions), len(newPerms))
            if *apply {
                if _, err := db.UpdateRoleDefinitionPermissions(ctx, rd.Id, newPerms); err != nil {
                    fmt.Fprintf(os.Stderr, "update role %s: %v\n", rd.Id, err)
                } else {
                    changedRoles++
                }
            }
        }
    }

    // Normalize ownership policies
    changedOwnership := 0
    for _, op := range ownerships {
        newOwner := normalizePermList(op.OwnerPermissions, defsMap)
        newParent := normalizePermList(op.ParentOwnerPermissions, defsMap)
        if !equalStringSlices(newOwner, op.OwnerPermissions) || !equalStringSlices(newParent, op.ParentOwnerPermissions) {
            fmt.Printf("OwnershipPolicy %s/%s will be updated\n", op.TenantId, op.ResourceType)
            if *apply {
                if _, err := db.UpsertOwnershipPolicy(ctx, op.TenantId, op.ResourceType, newOwner, newParent); err != nil {
                    fmt.Fprintf(os.Stderr, "upsert ownership %s/%s: %v\n", op.TenantId, op.ResourceType, err)
                } else {
                    changedOwnership++
                }
            }
        }
    }

    fmt.Printf("Dry-run=%v. Roles changed (applied): %d. Ownership policies changed (applied): %d\n", *apply, changedRoles, changedOwnership)
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
            // skip malformed
            out = append(out, p)
            continue
        }
        res := parts[0]
        act := strings.ToLower(parts[1])
        // Direct match
        if _, ok := defsMap[res]; ok {
            out = append(out, models.GetPermission(res, act))
            continue
        }
        // Try snake_case conversion and lowercasing
        candidate := toSnake(res)
        if _, ok := defsMap[candidate]; ok {
            out = append(out, models.GetPermission(candidate, act))
            continue
        }
        // Try lowercased resource
        lower := strings.ToLower(res)
        if _, ok := defsMap[lower]; ok {
            out = append(out, models.GetPermission(lower, act))
            continue
        }
        // Unknown resource: keep original but normalize action case
        out = append(out, fmt.Sprintf("%s:%s", res, act))
    }
    return out
}

func toSnake(s string) string {
    // Replace dashes and spaces
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

