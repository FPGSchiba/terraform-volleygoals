package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

var (
	sortByQueryNames    = []string{"sort_by", "sortBy"}
	sortOrderQueryNames = []string{"sort_order", "sortOrder"}
	limitQueryName      = "limit"
	nextTokenQueryNames = []string{"next_token", "nextToken"}
)

const (
	maxPageSize     = 100
	defaultPageSize = 25
)

// FilterOptions holds generic sorting & paging options reusable across resources.
type FilterOptions struct {
	SortBy    string         // e.g. "createdAt" | "name"
	SortOrder string         // "asc" | "desc"
	Limit     int            // page size
	Cursor    *models.Cursor // opaque cursor decoded into Cursor struct
}

func getFirst(keys []string, q map[string]string) (string, bool) {
	for _, k := range keys {
		if v, ok := q[k]; ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v, true
			}
		}
	}
	return "", false
}

// NormalizeSort returns canonical sort and order values.
func (f *FilterOptions) NormalizeSort() (string, string) {
	sortBy := strings.TrimSpace(strings.ToLower(f.SortBy))
	sortOrder := strings.TrimSpace(strings.ToLower(f.SortOrder))
	if sortBy == "" {
		return "", ""
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}
	return sortBy, sortOrder
}

// FilterOptionsFromQuery parses sorting and pagination parameters from QueryStringParameters.
// It supports both snake_case and camelCase keys: "sort_by" / "sortBy", "sort_order" / "sortOrder",
// "limit", and "next_token".
func FilterOptionsFromQuery(q map[string]string, defaultLimit, maxLimit int) (FilterOptions, error) {
	var f FilterOptions
	if q == nil {
		f.Limit = defaultLimit
		return f, nil
	}

	if v, ok := getFirst(sortByQueryNames, q); ok {
		f.SortBy = v
	}
	if v, ok := getFirst(sortOrderQueryNames, q); ok {
		f.SortOrder = v
	}

	if v, ok := q[limitQueryName]; ok {
		lim, err := models.ParseLimit(strings.TrimSpace(v), defaultLimit, maxLimit)
		if err != nil {
			return f, err
		}
		f.Limit = lim
	} else {
		f.Limit = defaultLimit
	}

	if tok, ok := getFirst(nextTokenQueryNames, q); ok {
		cur, err := models.DecodeCursor(tok)
		if err != nil {
			return f, err
		}
		f.Cursor = cur
	}

	return f, nil
}

// TeamFilter combines resource-specific filters for teams with generic sort & pagination options.
type TeamFilter struct {
	FilterOptions
	NameContains string // partial match against teamName
	Status       string // "active" | "inactive" | ""
}

// BuildExpression builds a DynamoDB filter expression for teams.
func (f *TeamFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.NameContains) != "" {
		parts = append(parts, "contains(teamName, :name)")
		values[":name"] = &types.AttributeValueMemberS{Value: f.NameContains}
	}
	if strings.TrimSpace(f.Status) != "" {
		parts = append(parts, "#s = :status")
		names["#s"] = "status"
		values[":status"] = &types.AttributeValueMemberS{Value: f.Status}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// TeamFilterFromQuery parses team-specific and generic filter params from QueryStringParameters.
// Returns an error if limit or cursor parsing fails.
func TeamFilterFromQuery(q map[string]string) (TeamFilter, error) {
	var t TeamFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return t, err
	}
	t.FilterOptions = fo

	// name / teamName
	if v, ok := q["name"]; ok && strings.TrimSpace(v) != "" {
		t.NameContains = strings.TrimSpace(v)
	}

	// status
	if v, ok := q["status"]; ok {
		t.Status = strings.TrimSpace(v)
	}

	return t, nil
}

// TeamInviteFilter combines resource-specific filters for teams with generic sort & pagination options.
type TeamInviteFilter struct {
	FilterOptions
	EmailContains string // partial match against email
	Status        string // "pending" | "accepted" | "declined" | "revoked" | "expired" | ""
	Role          string // "member" | "admin" | "trainer" | ""
	InvitedBy     string // userId of the inviter
}

// BuildExpression builds a DynamoDB filter expression for team invites.
func (f *TeamInviteFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.EmailContains) != "" {
		parts = append(parts, "contains(email, :email)")
		values[":email"] = &types.AttributeValueMemberS{Value: f.EmailContains}
	}
	if strings.TrimSpace(f.Status) != "" {
		parts = append(parts, "#s = :status")
		names["#s"] = "status"
		values[":status"] = &types.AttributeValueMemberS{Value: f.Status}
	}
	if strings.TrimSpace(f.Role) != "" {
		parts = append(parts, "#r = :role")
		names["#r"] = "role"
		values[":role"] = &types.AttributeValueMemberS{Value: f.Role}
	}
	if strings.TrimSpace(f.InvitedBy) != "" {
		parts = append(parts, "#ib = :invitedBy")
		names["#ib"] = "invitedBy"
		values[":invitedBy"] = &types.AttributeValueMemberS{Value: f.InvitedBy}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// TeamInviteFilterFromQuery parses team-invite-specific and generic filter params from QueryStringParameters.
// Returns an error if limit or cursor parsing fails.
func TeamInviteFilterFromQuery(q map[string]string) (TeamInviteFilter, error) {
	var t TeamInviteFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return t, err
	}
	t.FilterOptions = fo

	// email
	if v, ok := q["email"]; ok && strings.TrimSpace(v) != "" {
		t.EmailContains = strings.TrimSpace(v)
	}

	// status
	if v, ok := q["status"]; ok {
		t.Status = strings.TrimSpace(v)
	}

	// role
	if v, ok := q["role"]; ok {
		t.Role = strings.TrimSpace(v)
	}

	// invitedBy
	if v, ok := q["invitedBy"]; ok {
		t.InvitedBy = strings.TrimSpace(v)
	}

	return t, nil
}

type TeamMemberFilter struct {
	FilterOptions
	Role         string // "member" | "admin" | "trainer" | ""
	UserId       string // userId of the team member
	Status       string // "active" | "invited" | "removed" | "left" | ""
	NameContains  string // partial match on user name (applied in-memory after enrichment)
	EmailContains string // partial match on user email (applied in-memory after enrichment)
}

// BuildExpression builds a DynamoDB filter expression for team members.
func (f *TeamMemberFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.Role) != "" {
		parts = append(parts, "#r = :role")
		names["#r"] = "role"
		values[":role"] = &types.AttributeValueMemberS{Value: f.Role}
	}

	if strings.TrimSpace(f.UserId) != "" {
		parts = append(parts, "#u = :userId")
		names["#u"] = "userId"
		values[":userId"] = &types.AttributeValueMemberS{Value: f.UserId}
	}

	if strings.TrimSpace(f.Status) != "" {
		parts = append(parts, "#s = :status")
		names["#s"] = "status"
		values[":status"] = &types.AttributeValueMemberS{Value: f.Status}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// TeamMemberFilterFromQuery parses team-member-specific and generic filter params from QueryStringParameters.
// Returns an error if limit or cursor parsing fails.
func TeamMemberFilterFromQuery(q map[string]string) (TeamMemberFilter, error) {
	var t TeamMemberFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return t, err
	}
	t.FilterOptions = fo

	// role
	if v, ok := q["role"]; ok {
		t.Role = strings.TrimSpace(v)
	}

	// userId
	if v, ok := q["userId"]; ok {
		t.UserId = strings.TrimSpace(v)
	}

	// status
	if v, ok := q["status"]; ok {
		t.Status = strings.TrimSpace(v)
	}

	// name / name_contains — in-memory filter after user enrichment
	for _, key := range []string{"name", "name_contains"} {
		if v, ok := q[key]; ok && strings.TrimSpace(v) != "" {
			t.NameContains = strings.TrimSpace(v)
			break
		}
	}

	// email / email_contains — in-memory filter after user enrichment
	for _, key := range []string{"email", "email_contains"} {
		if v, ok := q[key]; ok && strings.TrimSpace(v) != "" {
			t.EmailContains = strings.TrimSpace(v)
			break
		}
	}

	return t, nil
}

// SeasonFilter combines season-specific filters with generic sort & pagination options.
type SeasonFilter struct {
	FilterOptions
	TeamId       string // exact match on teamId
	NameContains string // partial match against name
	Status       string // season status
}

// BuildExpression builds a DynamoDB filter expression for seasons.
func (f *SeasonFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.TeamId) != "" {
		parts = append(parts, "#teamId = :teamId")
		names["#teamId"] = "teamId"
		values[":teamId"] = &types.AttributeValueMemberS{Value: f.TeamId}
	}

	if strings.TrimSpace(f.NameContains) != "" {
		parts = append(parts, "contains(#n, :name)")
		names["#n"] = "name"
		values[":name"] = &types.AttributeValueMemberS{Value: f.NameContains}
	}

	if strings.TrimSpace(f.Status) != "" {
		parts = append(parts, "#s = :status")
		names["#s"] = "status"
		values[":status"] = &types.AttributeValueMemberS{Value: f.Status}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// SeasonFilterFromQuery parses season-specific and generic filter params from QueryStringParameters.
func SeasonFilterFromQuery(q map[string]string) (SeasonFilter, error) {
	var s SeasonFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return s, err
	}
	s.FilterOptions = fo

	if v, ok := q["teamId"]; ok {
		s.TeamId = strings.TrimSpace(v)
	}
	if v, ok := q["name"]; ok && strings.TrimSpace(v) != "" {
		s.NameContains = strings.TrimSpace(v)
	}
	if v, ok := q["status"]; ok {
		s.Status = strings.TrimSpace(v)
	}

	return s, nil
}

// GoalFilter combines goal-specific filters with generic sort & pagination options.
type GoalFilter struct {
	FilterOptions
	OwnerId       string   // exact match on ownerId
	GoalType      string   // goal type (individual|team)
	Status        string   // goal status
	TitleContains string   // partial match against title
	TeamId        string   // exact match on teamId
	GoalIds       []string // restrict to specific goal IDs (OR condition)
}

// BuildExpression builds a DynamoDB filter expression for goals.
func (f *GoalFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.OwnerId) != "" {
		parts = append(parts, "#ownerId = :ownerId")
		names["#ownerId"] = "ownerId"
		values[":ownerId"] = &types.AttributeValueMemberS{Value: f.OwnerId}
	}

	if strings.TrimSpace(f.GoalType) != "" {
		parts = append(parts, "#gt = :goalType")
		names["#gt"] = "goalType"
		values[":goalType"] = &types.AttributeValueMemberS{Value: f.GoalType}
	}

	if strings.TrimSpace(f.Status) != "" {
		parts = append(parts, "#s = :status")
		names["#s"] = "status"
		values[":status"] = &types.AttributeValueMemberS{Value: f.Status}
	}

	if strings.TrimSpace(f.TitleContains) != "" {
		parts = append(parts, "contains(#t, :title)")
		names["#t"] = "title"
		values[":title"] = &types.AttributeValueMemberS{Value: f.TitleContains}
	}

	if strings.TrimSpace(f.TeamId) != "" {
		parts = append(parts, "#teamId = :teamId")
		names["#teamId"] = "teamId"
		values[":teamId"] = &types.AttributeValueMemberS{Value: f.TeamId}
	}

	if len(f.GoalIds) > 0 {
		idParts := make([]string, 0, len(f.GoalIds))
		for i, id := range f.GoalIds {
			k := fmt.Sprintf(":gid%d", i)
			values[k] = &types.AttributeValueMemberS{Value: id}
			idParts = append(idParts, fmt.Sprintf("id = %s", k))
		}
		parts = append(parts, "("+strings.Join(idParts, " OR ")+")")
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// GoalFilterFromQuery parses goal-specific and generic filter params from QueryStringParameters.
func GoalFilterFromQuery(q map[string]string) (GoalFilter, error) {
	var g GoalFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return g, err
	}
	g.FilterOptions = fo

	if v, ok := q["ownerId"]; ok {
		g.OwnerId = strings.TrimSpace(v)
	}
	if v, ok := q["goalType"]; ok {
		g.GoalType = strings.TrimSpace(v)
	}
	if v, ok := q["status"]; ok {
		g.Status = strings.TrimSpace(v)
	}
	if v, ok := q["title"]; ok && strings.TrimSpace(v) != "" {
		g.TitleContains = strings.TrimSpace(v)
	}

	return g, nil
}

// ProgressReportFilter combines progress-report-specific filters with generic sort & pagination options.
type ProgressReportFilter struct {
	FilterOptions
	SeasonId        string     // exact match on seasonId
	AuthorId        string     // exact match on authorId
	SummaryContains string     // contains() match on summary
	CreatedAfter    *time.Time // createdAt >= CreatedAfter (inclusive)
	CreatedBefore   *time.Time // createdAt <= CreatedBefore (inclusive)
}

// BuildExpression builds a DynamoDB filter expression for progress reports.
func (f *ProgressReportFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.SeasonId) != "" {
		parts = append(parts, "#seasonId = :seasonId")
		names["#seasonId"] = "seasonId"
		values[":seasonId"] = &types.AttributeValueMemberS{Value: f.SeasonId}
	}

	if strings.TrimSpace(f.AuthorId) != "" {
		parts = append(parts, "#authorId = :authorId")
		names["#authorId"] = "authorId"
		values[":authorId"] = &types.AttributeValueMemberS{Value: f.AuthorId}
	}

	if strings.TrimSpace(f.SummaryContains) != "" {
		parts = append(parts, "contains(#summary, :summary)")
		names["#summary"] = "summary"
		values[":summary"] = &types.AttributeValueMemberS{Value: f.SummaryContains}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// ProgressReportFilterFromQuery parses progress-report-specific and generic filter params from QueryStringParameters.
func ProgressReportFilterFromQuery(q map[string]string) (ProgressReportFilter, error) {
	var p ProgressReportFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return p, err
	}
	p.FilterOptions = fo

	if v, ok := q["seasonId"]; ok {
		p.SeasonId = strings.TrimSpace(v)
	}
	if v, ok := q["authorId"]; ok {
		p.AuthorId = strings.TrimSpace(v)
	}
	if v, ok := q["summary"]; ok && strings.TrimSpace(v) != "" {
		p.SummaryContains = strings.TrimSpace(v)
	}

	if v, ok := q["createdAfter"]; ok && strings.TrimSpace(v) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(v))
		if err != nil {
			return p, fmt.Errorf("invalid createdAfter: must be RFC3339 / ISO 8601")
		}
		p.CreatedAfter = &t
	}

	if v, ok := q["createdBefore"]; ok && strings.TrimSpace(v) != "" {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(v))
		if err != nil {
			return p, fmt.Errorf("invalid createdBefore: must be RFC3339 / ISO 8601")
		}
		p.CreatedBefore = &t
	}

	return p, nil
}

// ActivityFilter combines activity-specific filters with generic sort & pagination options.
type ActivityFilter struct {
	FilterOptions
	TeamId string // exact match on teamId
}

// BuildExpression builds a DynamoDB filter expression for activities.
func (f *ActivityFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.TeamId) != "" {
		parts = append(parts, "#teamId = :teamId")
		names["#teamId"] = "teamId"
		values[":teamId"] = &types.AttributeValueMemberS{Value: f.TeamId}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// ActivityFilterFromQuery parses activity-specific and generic filter params from QueryStringParameters.
func ActivityFilterFromQuery(q map[string]string) (ActivityFilter, error) {
	var a ActivityFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return a, err
	}
	a.FilterOptions = fo

	return a, nil
}

// CommentFilter combines comment-specific filters with generic sort & pagination options.
// TargetId and CommentType are required.
type CommentFilter struct {
	FilterOptions
	TargetId    string // required exact match on targetId
	CommentType string // required exact match on commentType ("Goal" | "ProgressReport" | "ProgressEntry")
	AuthorId    string // optional exact match on authorId
}

// BuildExpression builds a DynamoDB filter expression for comments.
func (f *CommentFilter) BuildExpression() (string, map[string]types.AttributeValue, map[string]string) {
	parts := make([]string, 0)
	values := make(map[string]types.AttributeValue)
	names := make(map[string]string)

	if strings.TrimSpace(f.TargetId) != "" {
		parts = append(parts, "#targetId = :targetId")
		names["#targetId"] = "targetId"
		values[":targetId"] = &types.AttributeValueMemberS{Value: f.TargetId}
	}

	if strings.TrimSpace(f.CommentType) != "" {
		parts = append(parts, "#commentType = :commentType")
		names["#commentType"] = "commentType"
		values[":commentType"] = &types.AttributeValueMemberS{Value: f.CommentType}
	}

	if strings.TrimSpace(f.AuthorId) != "" {
		parts = append(parts, "#authorId = :authorId")
		names["#authorId"] = "authorId"
		values[":authorId"] = &types.AttributeValueMemberS{Value: f.AuthorId}
	}

	if len(parts) == 0 {
		return "", nil, nil
	}
	return strings.Join(parts, " AND "), values, names
}

// CommentFilterFromQuery parses comment-specific and generic filter params from QueryStringParameters.
// Returns an error if targetId or commentType are missing.
func CommentFilterFromQuery(q map[string]string) (CommentFilter, error) {
	var c CommentFilter

	fo, err := FilterOptionsFromQuery(q, defaultPageSize, maxPageSize)
	if err != nil {
		return c, err
	}
	c.FilterOptions = fo

	targetId, ok := q["targetId"]
	if !ok || strings.TrimSpace(targetId) == "" {
		return c, fmt.Errorf("targetId is required")
	}
	c.TargetId = strings.TrimSpace(targetId)

	commentType, ok := q["commentType"]
	if !ok || strings.TrimSpace(commentType) == "" {
		return c, fmt.Errorf("commentType is required")
	}
	c.CommentType = strings.TrimSpace(commentType)

	if v, ok := q["authorId"]; ok {
		c.AuthorId = strings.TrimSpace(v)
	}

	return c, nil
}
