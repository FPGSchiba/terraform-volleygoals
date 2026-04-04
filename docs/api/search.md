# Search & Health API

---

## Global Search

### Search

- **Endpoint**: `GET /api/v1/search`
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

## Health Check

### Health Check

- **Endpoint**: `GET /health`
- **Auth**: **Public (no auth required)**
- **Description**: Checks the health of all service dependencies (DynamoDB, S3, Cognito). Returns `200` if all healthy, `503` if any dependency is unhealthy.

> **Note**: This endpoint is at the root path (`/health`), not under `/api/v1`.

**Response** `200 OK` (all healthy)
```json
{
  "message": "Success",
  "status": "healthy",
  "dependencies": {
    "dynamodb": "ok",
    "s3": "ok",
    "cognito": "ok"
  }
}
```

**Response** `503 Service Unavailable` (unhealthy)
```json
{
  "message": "Success",
  "status": "unhealthy",
  "dependencies": {
    "dynamodb": "error",
    "s3": "ok",
    "cognito": "ok"
  }
}
```

| Field | Value | Description |
|-------|-------|-------------|
| `status` | `"healthy"` / `"unhealthy"` | Overall service health |
| `dependencies.dynamodb` | `"ok"` / `"error"` | DynamoDB connectivity |
| `dependencies.s3` | `"ok"` / `"error"` | S3 connectivity |
| `dependencies.cognito` | `"ok"` / `"error"` | Cognito connectivity |
