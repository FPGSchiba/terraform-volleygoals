package db

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

// ProgressEntry is a goal rating within a progress report.
type ProgressEntry struct {
	GoalId string
	Rating int8
}

func CreateProgressReport(ctx context.Context, seasonId, authorId, summary, details string, progressEntries []ProgressEntry) (*models.ProgressReport, error) {
	client = GetClient()
	now := time.Now()
	report := &models.ProgressReport{
		Id:        models.GenerateID(),
		SeasonId:  seasonId,
		AuthorId:  authorId,
		Summary:   summary,
		Details:   details,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &progressReportsTableName,
		Item:      report.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}

	if err := writeProgressEntries(ctx, report.Id, progressEntries); err != nil {
		return nil, err
	}

	return report, nil
}

func GetProgressReportById(ctx context.Context, reportId string) (*models.ProgressReport, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &progressReportsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: reportId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var report models.ProgressReport
	err = attributevalue.UnmarshalMap(result.Item, &report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func UpdateProgressReport(ctx context.Context, reportId string, summary, details *string, progressEntries []ProgressEntry) (*models.ProgressReport, error) {
	client = GetClient()
	updateParts := []string{}
	exprAttrValues := make(map[string]types.AttributeValue)
	exprAttrNames := make(map[string]string)

	if summary != nil {
		updateParts = append(updateParts, "#summary = :summary")
		exprAttrNames["#summary"] = "summary"
		exprAttrValues[":summary"] = &types.AttributeValueMemberS{Value: *summary}
	}

	if details != nil {
		updateParts = append(updateParts, "#details = :details")
		exprAttrNames["#details"] = "details"
		exprAttrValues[":details"] = &types.AttributeValueMemberS{Value: *details}
	}

	updateParts = append(updateParts, "#updatedAt = :updatedAt")
	exprAttrNames["#updatedAt"] = "updatedAt"
	exprAttrValues[":updatedAt"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}

	updateExpr := "SET " + strings.Join(updateParts, ", ")

	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &progressReportsTableName,
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: reportId}},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}

	var updatedReport models.ProgressReport
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedReport)
	if err != nil {
		return nil, err
	}

	if progressEntries != nil {
		if err := deleteProgressEntriesByReportId(ctx, reportId); err != nil {
			return nil, err
		}
		if err := writeProgressEntries(ctx, reportId, progressEntries); err != nil {
			return nil, err
		}
	}

	return &updatedReport, nil
}

func DeleteProgressReport(ctx context.Context, reportId string) error {
	client = GetClient()
	if err := deleteProgressEntriesByReportId(ctx, reportId); err != nil {
		return err
	}
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &progressReportsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: reportId},
		},
	})
	return err
}

func ListProgressReports(ctx context.Context, filter ProgressReportFilter) ([]*models.ProgressReport, int, *models.Cursor, bool, error) {
	client = GetClient()

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	in := &dynamodb.ScanInput{
		TableName: aws.String(progressReportsTableName),
		Limit:     aws.Int32(int32(limit)),
	}

	expr, vals, names := filter.BuildExpression()
	if strings.TrimSpace(expr) != "" {
		in.FilterExpression = aws.String(expr)
		in.ExpressionAttributeValues = vals
		if len(names) > 0 {
			in.ExpressionAttributeNames = names
		}
	}

	if filter.Cursor != nil && filter.Cursor.LastID != "" {
		in.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: filter.Cursor.LastID},
		}
	}

	result, err := client.Scan(ctx, in)
	if err != nil {
		return nil, 0, nil, false, err
	}

	reports, err := unmarshalProgressReports(result.Items)
	if err != nil {
		return nil, 0, nil, false, err
	}

	if sortBy, sortOrder := filter.NormalizeSort(); sortBy != "" {
		sortProgressReports(reports, sortBy, sortOrder)
	}

	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)

	return reports, len(reports), nextCursor, hasMore, nil
}

func unmarshalProgressReports(items []map[string]types.AttributeValue) ([]*models.ProgressReport, error) {
	reports := make([]*models.ProgressReport, 0, len(items))
	for _, it := range items {
		var r models.ProgressReport
		if err := attributevalue.UnmarshalMap(it, &r); err != nil {
			return nil, err
		}
		reports = append(reports, &r)
	}
	return reports, nil
}

func sortProgressReports(reports []*models.ProgressReport, sortBy, sortOrder string) {
	desc := strings.ToLower(sortOrder) == "desc"
	switch strings.ToLower(sortBy) {
	case "createdat", "created_at":
		sort.Slice(reports, func(i, j int) bool {
			if desc {
				return reports[i].CreatedAt.After(reports[j].CreatedAt)
			}
			return reports[i].CreatedAt.Before(reports[j].CreatedAt)
		})
	case "updatedat", "updated_at":
		sort.Slice(reports, func(i, j int) bool {
			if desc {
				return reports[i].UpdatedAt.After(reports[j].UpdatedAt)
			}
			return reports[i].UpdatedAt.Before(reports[j].UpdatedAt)
		})
	}
}

func writeProgressEntries(ctx context.Context, reportId string, entries []ProgressEntry) error {
	for _, e := range entries {
		progress := &models.Progress{
			Id:               models.GenerateID(),
			ProgressReportId: reportId,
			GoalId:           e.GoalId,
			Rating:           e.Rating,
		}
		_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &progressTableName,
			Item:      progress.ToAttributeValues(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteProgressEntriesByReportId(ctx context.Context, reportId string) error {
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(progressTableName),
		FilterExpression: aws.String("#rid = :reportId"),
		ExpressionAttributeNames: map[string]string{
			"#rid": "progressReportId",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":reportId": &types.AttributeValueMemberS{Value: reportId},
		},
	})
	if err != nil {
		return err
	}
	for _, item := range result.Items {
		var p models.Progress
		if err := attributevalue.UnmarshalMap(item, &p); err != nil {
			return err
		}
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &progressTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: p.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
