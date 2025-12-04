package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
	log "github.com/sirupsen/logrus"
)

func GetTeamMemberByUserIDAndTeamID(ctx context.Context, userID string, teamID string) (*models.TeamMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName: &teamMembersTableName,
		IndexName: aws.String("teamUserIdIndex"),
		KeyConditions: map[string]types.Condition{
			"userId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: userID},
				},
			},
			"teamId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: teamID},
				},
			},
		},
		FilterExpression: aws.String("#status = :activeStatus"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":activeStatus": &types.AttributeValueMemberS{Value: string(models.TeamMemberStatusActive)},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var teamMember models.TeamMember
	err = attributevalue.UnmarshalMap(result.Items[0], &teamMember)
	if err != nil {
		return nil, err
	}
	return &teamMember, nil
}

func HasRoleOnTeam(ctx context.Context, userID string, teamID string, role models.TeamMemberRole) (bool, error) {
	teamMember, err := GetTeamMemberByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return false, err
	}
	if teamMember != nil && teamMember.Role == role {
		return true, nil
	}
	return false, nil
}

func GetMembershipsByUserID(ctx context.Context, userID string) ([]*models.TeamMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName: &teamMembersTableName,
		IndexName: aws.String("userIdIndex"),
		KeyConditions: map[string]types.Condition{
			"userId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: userID},
				},
			},
		},
		FilterExpression: aws.String("#status = :activeStatus"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":activeStatus": &types.AttributeValueMemberS{Value: string(models.TeamMemberStatusActive)},
		},
	})
	if err != nil {
		return nil, err
	}
	var teamMembers []*models.TeamMember
	err = attributevalue.UnmarshalListOfMaps(result.Items, &teamMembers)
	if err != nil {
		return nil, err
	}
	return teamMembers, nil
}

func DeleteTeamMembershipsByTeamID(ctx context.Context, teamID string) error {
	memberships, err := GetMembershipsByTeamID(ctx, teamID)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &teamMembersTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: membership.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteTeamMembershipsByUserID(ctx context.Context, userID string) error {
	memberships, err := GetMembershipsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &teamMembersTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: membership.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetMembershipsByTeamID(ctx context.Context, teamID string) ([]*models.TeamMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName: &teamMembersTableName,
		IndexName: aws.String("teamIdIndex"),
		KeyConditions: map[string]types.Condition{
			"teamId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: teamID},
				},
			},
		},
		FilterExpression: aws.String("#status = :activeStatus"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":activeStatus": &types.AttributeValueMemberS{Value: string(models.TeamMemberStatusActive)},
		},
	})
	if err != nil {
		return nil, err
	}
	var teamMembers []*models.TeamMember
	err = attributevalue.UnmarshalListOfMaps(result.Items, &teamMembers)
	if err != nil {
		return nil, err
	}
	return teamMembers, nil
}

func CreateTeamMemberFromInvite(ctx context.Context, invite *models.Invite) (*models.TeamMember, error) {
	log.Printf("[DEBUG] CreateTeamMemberFromInvite AcceptedBy: %v", invite.AcceptedBy)
	client = GetClient()
	timeNow := time.Now()
	if invite.AcceptedBy == nil {
		return nil, fmt.Errorf("CreateTeamMemberFromInvite: invite.AcceptedBy is nil")
	}
	teamMember := &models.TeamMember{
		Id:        models.GenerateID(),
		TeamId:    invite.TeamId,
		UserId:    *invite.AcceptedBy,
		Role:      invite.Role,
		Status:    models.TeamMemberStatusActive,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		JoinedAt:  &timeNow,
	}
	item, err := attributevalue.MarshalMap(teamMember)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &teamMembersTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return teamMember, nil
}

func IsUserMemberOfTeam(ctx context.Context, userID string, teamID string) (bool, error) {
	teamMember, err := GetTeamMemberByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return false, err
	}
	return teamMember != nil, nil
}

func GetTeamAssignmentsByUserID(ctx context.Context, userID string) ([]*models.TeamAssignment, error) {
	teamMembers, err := GetMembershipsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var teamAssignments []*models.TeamAssignment
	for _, tm := range teamMembers {
		team, err := GetTeamById(ctx, tm.TeamId)
		if err != nil {
			return nil, err
		}
		if team == nil {
			log.Printf("[WARN] GetTeamAssignmentsByUserID: team not found for teamId %s", tm.TeamId)
			continue
		}
		teamAssignment := &models.TeamAssignment{
			Team:   *team,
			Role:   tm.Role,
			Status: tm.Status,
		}
		teamAssignments = append(teamAssignments, teamAssignment)
	}
	return teamAssignments, nil
}

// ListTeamMembers returns a page of team members according to TeamMemberFilter.
func ListTeamMembers(ctx context.Context, teamId string, filter TeamMemberFilter) ([]*models.TeamMember, int, *models.Cursor, bool, error) {
	client = GetClient()

	// ensure sane default limit
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	in := &dynamodb.ScanInput{
		TableName: &teamMembersTableName,
		Limit:     aws.Int32(int32(limit)),
	}

	// Build filter expression using TeamMemberFilter
	if expr, vals, names := filter.BuildExpression(); expr != "" {
		in.FilterExpression = aws.String(expr)
		in.ExpressionAttributeValues = vals
		if len(names) > 0 {
			in.ExpressionAttributeNames = names
		}
	}

	// Always filter by teamId
	teamIdFilter := "#teamId = :teamId"
	if in.FilterExpression != nil && *in.FilterExpression != "" {
		*in.FilterExpression = *in.FilterExpression + " AND " + teamIdFilter
	} else {
		in.FilterExpression = aws.String(teamIdFilter)
	}
	if in.ExpressionAttributeNames == nil {
		in.ExpressionAttributeNames = make(map[string]string)
	}
	in.ExpressionAttributeNames["#teamId"] = "teamId"
	if in.ExpressionAttributeValues == nil {
		in.ExpressionAttributeValues = make(map[string]types.AttributeValue)
	}
	in.ExpressionAttributeValues[":teamId"] = &types.AttributeValueMemberS{Value: teamId}

	// Resume from cursor if provided
	if filter.Cursor != nil && filter.Cursor.LastID != "" {
		in.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: filter.Cursor.LastID},
		}
	}

	result, err := client.Scan(ctx, in)
	if err != nil {
		return nil, 0, nil, false, err
	}

	var members []*models.TeamMember
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &members); err != nil {
		return nil, 0, nil, false, err
	}

	// Sorting (in-memory) using embedded FilterOptions
	if sortBy, sortOrder := filter.NormalizeSort(); sortBy != "" {
		sortTeamMembers(members, sortBy, sortOrder)
	}

	// Build cursor from LastEvaluatedKey
	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)

	return members, len(members), nextCursor, hasMore, nil
}

// sortTeamMembers sorts the slice of TeamMember according to sortBy and sortOrder.
// Supported sortBy values: "createdAt", "joinedAt", "role", "userId"
func sortTeamMembers(members []*models.TeamMember, sortBy, sortOrder string) {
	desc := strings.ToLower(sortOrder) == "desc"
	switch strings.ToLower(sortBy) {
	case "joinedat", "joined_at":
		sortByJoinedAt(members, desc)
	case "createdat", "created_at":
		sortByCreatedAt(members, desc)
	case "role":
		sortByRole(members, desc)
	case "userid", "user_id":
		sortByUserID(members, desc)
	}
}

func sortByJoinedAt(members []*models.TeamMember, desc bool) {
	sort.Slice(members, func(i, j int) bool {
		a := members[i].JoinedAt
		b := members[j].JoinedAt
		var at time.Time
		if a != nil {
			at = *a
		}
		var bt time.Time
		if b != nil {
			bt = *b
		}
		if desc {
			return at.After(bt)
		}
		return at.Before(bt)
	})
}

func sortByCreatedAt(members []*models.TeamMember, desc bool) {
	sort.Slice(members, func(i, j int) bool {
		if desc {
			return members[i].CreatedAt.After(members[j].CreatedAt)
		}
		return members[i].CreatedAt.Before(members[j].CreatedAt)
	})
}

func sortByRole(members []*models.TeamMember, desc bool) {
	sort.Slice(members, func(i, j int) bool {
		if desc {
			return members[i].Role > members[j].Role
		}
		return members[i].Role < members[j].Role
	})
}

func sortByUserID(members []*models.TeamMember, desc bool) {
	sort.Slice(members, func(i, j int) bool {
		if desc {
			return members[i].UserId > members[j].UserId
		}
		return members[i].UserId < members[j].UserId
	})
}

func AddTeamMember(ctx context.Context, teamId, userId string, role models.TeamMemberRole) (*models.TeamMember, error) {
	client = GetClient()
	timeNow := time.Now()
	teamMember := &models.TeamMember{
		Id:        models.GenerateID(),
		TeamId:    teamId,
		UserId:    userId,
		Role:      role,
		Status:    models.TeamMemberStatusActive,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		JoinedAt:  &timeNow,
	}
	item, err := attributevalue.MarshalMap(teamMember)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &teamMembersTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return teamMember, nil
}

func UpdateTeamMember(ctx context.Context, teamMemberId string, role *models.TeamMemberRole, status *models.TeamMemberStatus) (*models.TeamMember, error) {
	client = GetClient()
	var updateExpressions []string
	exprAttrValues := make(map[string]types.AttributeValue)
	exprAttrNames := make(map[string]string)

	if role != nil {
		updateExpressions = append(updateExpressions, "#role = :role")
		exprAttrValues[":role"] = &types.AttributeValueMemberS{Value: string(*role)}
		exprAttrNames["#role"] = "role"
	}

	if status != nil {
		updateExpressions = append(updateExpressions, "#status = :status")
		exprAttrValues[":status"] = &types.AttributeValueMemberS{Value: string(*status)}
		exprAttrNames["#status"] = "status"
	}

	updateExpressions = append(updateExpressions, "#updatedAt = :updatedAt")
	exprAttrValues[":updatedAt"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}
	exprAttrNames["#updatedAt"] = "updatedAt"

	updateExpr := "SET " + strings.Join(updateExpressions, ", ")

	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &teamMembersTableName,
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: teamMemberId}},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}
	var updatedTeamMember models.TeamMember
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedTeamMember)
	if err != nil {
		return nil, err
	}
	return &updatedTeamMember, nil
}

func GetUserRoleOnTeam(ctx context.Context, userID string, teamID string) (*models.TeamMemberRole, error) {
	teamMember, err := GetTeamMemberByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	if teamMember == nil {
		return nil, nil
	}
	return &teamMember.Role, nil
}

func HasOtherAdminOrTrainer(ctx context.Context, teamID string, excludingUserID string) (bool, error) {
	teamMembers, err := GetMembershipsByTeamID(ctx, teamID)
	if err != nil {
		return false, err
	}
	for _, member := range teamMembers {
		if member.UserId != excludingUserID && (member.Role == models.TeamMemberRoleAdmin || member.Role == models.TeamMemberRoleTrainer) && member.Status == models.TeamMemberStatusActive {
			return true, nil
		}
	}
	return false, nil
}

func LeaveTeam(ctx context.Context, teamID string, userID string) error {
	teamMember, err := GetTeamMemberByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return err
	}
	if teamMember == nil {
		return fmt.Errorf("LeaveTeam: user %s is not a member of team %s", userID, teamID)
	}
	// Set status to left and update UpdatedAt and LeftAt
	timeNow := time.Now()
	updateExpr := "SET #status = :status, #updatedAt = :updatedAt, #leftAt = :leftAt"
	_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &teamMembersTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: teamMember.Id},
		},
		UpdateExpression: aws.String(updateExpr),
		ExpressionAttributeNames: map[string]string{
			"#status":    "status",
			"#updatedAt": "updatedAt",
			"#leftAt":    "leftAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":    &types.AttributeValueMemberS{Value: string(models.TeamMemberStatusLeft)},
			":updatedAt": &types.AttributeValueMemberS{Value: timeNow.Format(time.RFC3339)},
			":leftAt":    &types.AttributeValueMemberS{Value: timeNow.Format(time.RFC3339)},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func RemoveTeamMember(ctx context.Context, teamMemberId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &teamMembersTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: teamMemberId},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func RemoveAllTeamMembersForUser(ctx context.Context, userID string) error {
	memberships, err := GetMembershipsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &teamMembersTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: membership.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}
