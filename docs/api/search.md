# Search & Health API

---

## Global Search

### Search

- **Endpoint**: `GET /v1/search`
- **Auth**: Any authenticated team member with access to the specified team
- **Description**: Full-text search across goals and progress reports within a team. Results are ranked by relevance and capped at a configurable limit.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `q` | string | **Yes** | Search query string (must not be empty) |
| `teamId` | string (UUID) | **Yes** | Scope the search to this team |
| `limit` | integer | No | Max results to return (default: 10, max: 50) |
| `types` | string | No | Comma-separated list of types to search: `goals`, `reports`. Omit to search all. |

**Response** `200 OK`
```json
{
  "message": "Success",
  "data": {
    "results": [
      {
        "type": "goal",
        "id": "goal-uuid",
        "title": "Improve serve accuracy",
        "seasonId": "season-uuid",
        "status": "in_progress"
      },
      {
        "type": "report",
        "id": "report-uuid",
        "summary": "Good progress this week",
        "seasonId": "season-uuid",
        "createdAt": "2024-10-01T09:00:00Z"
      }
    ]
  }
}
```

**Goal Result Fields**

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always `"goal"` |
| `id` | string (UUID) | Goal ID |
| `title` | string | Goal title |
| `seasonId` | string (UUID) | Season the goal belongs to |
| `status` | string | Goal status |

**Report Result Fields**

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Always `"report"` |
| `id` | string (UUID) | Report ID |
| `summary` | string | Report summary |
| `seasonId` | string (UUID) | Season the report belongs to |
| `createdAt` | string (ISO 8601) | When the report was created |

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `q` or `teamId` |
| `403 Forbidden` | Not authorized to access this team |

---

## Activity Feed

### Get Team Activity

- **Endpoint**: `GET /v1/teams/{teamId}/activity`
- **Auth**: Global Admin or any team member
- **Description**: Returns a paginated activity feed for a team. Regular members (`role=member`) only receive events with `visibility=all`; admins and trainers also receive staff-only events.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "data": {
    "items": [
      {
        "id": "activity-uuid",
        "teamId": "team-uuid",
        "actorId": "cognito-sub",
        "actorName": "Jane Doe",
        "actorPicture": "https://cdn.example.com/users/jane.jpg",
        "action": "goal.status_changed",
        "description": "Goal \"Improve serve accuracy\" status changed to completed",
        "targetType": "goal",
        "targetId": "goal-uuid",
        "visibility": "all",
        "timestamp": "2024-11-01T12:00:00Z"
      }
    ],
    "count": 1,
    "nextToken": "",
    "hasMore": false
  }
}
```

**Activity Actions**

| Action | Visibility | Description |
|--------|------------|-------------|
| `goal.status_changed` | `all` | A goal's status was updated |
| `progress_report.created` | `all` | A new progress report was submitted |
| `member.joined` | `all` | A new member joined the team |
| `member.role_changed` | `admin_trainer` | A member's role was changed |
| `member.removed` | `admin_trainer` | A member was removed |
| `team_settings.updated` | `admin_trainer` | Team settings were changed |

**Visibility Values**

| Value | Who sees it |
|-------|-------------|
| `all` | All team members |
| `admin_trainer` | Admins and trainers only |

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId` |
| `403 Forbidden` | Not authorized |

---

## Health Check

### Health Check

- **Endpoint**: `POST /health`
- **Auth**: **Public (no auth required)**
- **Description**: Checks the health of all service dependencies (DynamoDB, S3, Cognito). Returns `200` if all healthy, `503` if any dependency is unhealthy.

> **Note**: This endpoint is at the root path (`/health`), not under `/v1`.

**Response** `200 OK` (all healthy)
```json
{
  "message": "Success",
  "data": {
    "status": "healthy",
    "dependencies": {
      "dynamodb": "ok",
      "s3": "ok",
      "cognito": "ok"
    }
  }
}
```

**Response** `503 Service Unavailable` (unhealthy)
```json
{
  "message": "Success",
  "data": {
    "status": "unhealthy",
    "dependencies": {
      "dynamodb": "error",
      "s3": "ok",
      "cognito": "ok"
    }
  }
}
```

| Field | Value | Description |
|-------|-------|-------------|
| `status` | `"healthy"` / `"unhealthy"` | Overall service health |
| `dependencies.dynamodb` | `"ok"` / `"error"` | DynamoDB connectivity |
| `dependencies.s3` | `"ok"` / `"error"` | S3 connectivity |
| `dependencies.cognito` | `"ok"` / `"error"` | Cognito connectivity |
