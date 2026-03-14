package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// CheckHealth verifies DynamoDB connectivity by describing the teams table.
func CheckHealth(ctx context.Context) error {
	client = GetClient()
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(teamsTableName),
	})
	return err
}
