# VolleyGoals API Documentation

This folder contains comprehensive API documentation for all VolleyGoals endpoints.

## Base URL

All versioned endpoints are prefixed with `/v1`. The health check lives at the root.

```
https://<api-id>.execute-api.<region>.amazonaws.com/<stage>/v1/...
```

## Authentication

All endpoints use **AWS Cognito User Pools** for authorization via a JWT Bearer token, except where marked **Public (no auth)**.

Include the token in the `Authorization` header:
```
Authorization: Bearer <id_token>
```

## Authorization Roles

| Scope | Description |
|-------|-------------|
| **Global Admin** | Full access to all resources |
| **Tenant Admin** | Admin of a specific tenant (organization) |
| **Team Admin** | Admin of a specific team |
| **Team Trainer** | Can manage goals, reports, and members |
| **Team Member** | Read-only access to shared resources |

## Common Response Envelope

All responses follow a consistent JSON envelope:

```json
{
  "message": "Success",
  "data": { ... }
}
```

### Paginated Responses

List endpoints return:
```json
{
  "message": "Success",
  "data": {
    "items": [...],
    "count": 25,
    "nextToken": "base64-encoded-cursor",
    "hasMore": true
  }
}
```

Pass `nextToken` as a query parameter to fetch the next page.

## Common Error Responses

| Status | Meaning |
|--------|---------|
| `400 Bad Request` | Missing or invalid request parameters/body |
| `401 Unauthorized` | Missing or invalid auth token |
| `403 Forbidden` | Authenticated but not authorized |
| `404 Not Found` | Resource does not exist |
| `409 Conflict` | Resource already exists |
| `500 Internal Server Error` | Unexpected server error |
| `503 Service Unavailable` | One or more dependencies are unhealthy |

## Documentation Index

| File | Resources |
|------|-----------|
| [teams.md](./teams.md) | Teams, Team Settings, Team Members |
| [seasons.md](./seasons.md) | Seasons |
| [goals.md](./goals.md) | Goals (nested under seasons) |
| [progress-reports.md](./progress-reports.md) | Progress Reports & Progress Entries |
| [comments.md](./comments.md) | Comments & Comment Files |
| [invites.md](./invites.md) | Team Invitations |
| [users.md](./users.md) | Users (admin-only) |
| [self.md](./self.md) | Current authenticated user |
| [tenants.md](./tenants.md) | Tenants, Roles, Ownership Policies |
| [search.md](./search.md) | Global Search & Health Check |
