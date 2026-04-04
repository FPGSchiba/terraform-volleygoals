# VolleyGoals API Documentation

This folder contains comprehensive API documentation for all VolleyGoals endpoints.

## Base URL

All versioned endpoints are prefixed with `/api/v1`. The health check lives at the root.

```
https://<api-id>.execute-api.<region>.amazonaws.com/<stage>/api/v1/...
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

Success responses are documented with a top-level `message` field plus resource-specific fields. When the payload is a map, its fields are flattened into the top-level response. When the payload is a non-map value such as an array, it is returned under `data`.

### Paginated Responses

List endpoints return the list fields directly:
```json
{
  "message": "Success",
  "items": [],
  "count": 25,
  "nextToken": "base64-encoded-cursor",
  "hasMore": true
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
| [teams.md](./teams.md) | Teams, Team Settings, Team Members, Team Activity |
| [seasons.md](./seasons.md) | Seasons |
| [goals.md](./goals.md) | Goals (nested under seasons) |
| [progress-reports.md](./progress-reports.md) | Progress Reports & Progress Entries |
| [comments.md](./comments.md) | Comments & Comment Files |
| [invites.md](./invites.md) | Team Invitations |
| [users.md](./users.md) | Users (admin-only) |
| [self.md](./self.md) | Current authenticated user |
| [tenants.md](./tenants.md) | Tenants, Tenant Members, Tenanted Teams, Resource Definitions, Resource Model, Roles, Ownership Policies |
| [search.md](./search.md) | Global Search & Health Check |
