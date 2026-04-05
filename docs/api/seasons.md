# Seasons API

All season endpoints are under `/api/v1/seasons`. Seasons belong to a team and act as the top-level container for Goals and Progress Reports.

---

### Create Season

- **Endpoint**: `POST /api/v1/seasons`
- **Auth**: Team member with `seasons:write` permission
- **Description**: Creates a new season for the given team.

**Request Body**
```json
{
  "teamId": "team-uuid",
  "name": "Season 2024/2025",
  "startDate": "2024-09-01T00:00:00Z",
  "endDate": "2025-05-31T23:59:59Z"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | Team this season belongs to |
| `name` | string | Yes | Display name of the season |
| `startDate` | string (ISO 8601) | Yes | Season start date/time |
| `endDate` | string (ISO 8601) | Yes | Season end date/time |

**Response** `201 Created`
```json
{
  "message": "Success",
  "season": {
    "id": "season-uuid",
    "teamId": "team-uuid",
    "name": "Season 2024/2025",
    "startDate": "2024-09-01T00:00:00Z",
    "endDate": "2025-05-31T23:59:59Z",
    "status": "planned",
    "createdAt": "2024-08-01T10:00:00Z",
    "updatedAt": "2024-08-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid or missing body fields |
| `403 Forbidden` | Not authorized |

---

### Get Season

- **Endpoint**: `GET /api/v1/seasons/{seasonId}`
- **Auth**: Team member with `seasons:read` permission (team member or above)
- **Description**: Returns a single season by ID.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | ID of the season |

**Response** `200 OK`
```json
{
  "message": "Success",
  "season": {
    "id": "season-uuid",
    "teamId": "team-uuid",
    "name": "Season 2024/2025",
    "startDate": "2024-09-01T00:00:00Z",
    "endDate": "2025-05-31T23:59:59Z",
    "status": "active",
    "createdAt": "2024-08-01T10:00:00Z",
    "updatedAt": "2024-09-01T00:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |

---

### List Seasons

- **Endpoint**: `GET /api/v1/seasons`
- **Auth**: Team member with `seasons:read` permission
- **Description**: Returns a paginated list of seasons for a specific team. `teamId` is required.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | **Yes** | Filter by team |
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "season-uuid",
      "teamId": "team-uuid",
      "name": "Season 2024/2025",
      "startDate": "2024-09-01T00:00:00Z",
      "endDate": "2025-05-31T23:59:59Z",
      "status": "active",
      "createdAt": "2024-08-01T10:00:00Z",
      "updatedAt": "2024-09-01T00:00:00Z"
    }
  ],
  "count": 1,
  "nextToken": "",
  "hasMore": false
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId` |
| `403 Forbidden` | Not authorized or `teamId` not provided |

---

### Update Season

- **Endpoint**: `PATCH /api/v1/seasons/{seasonId}`
- **Auth**: Team member with `seasons:write` permission (trainer or admin)
- **Description**: Updates a season's name, dates, or status.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | ID of the season |

**Request Body** (all fields optional)
```json
{
  "name": "Updated Season Name",
  "startDate": "2024-09-15T00:00:00Z",
  "endDate": "2025-06-15T23:59:59Z",
  "status": "active"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | New display name |
| `startDate` | string (ISO 8601) | No | New start date |
| `endDate` | string (ISO 8601) | No | New end date |
| `status` | string | No | `"planned"`, `"active"`, `"completed"`, or `"archived"` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "season": {
    "id": "season-uuid",
    "teamId": "team-uuid",
    "name": "Updated Season Name",
    "startDate": "2024-09-15T00:00:00Z",
    "endDate": "2025-06-15T23:59:59Z",
    "status": "active",
    "createdAt": "2024-08-01T10:00:00Z",
    "updatedAt": "2024-09-15T00:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |

---

### Delete Season

- **Endpoint**: `DELETE /api/v1/seasons/{seasonId}`
- **Auth**: Team member with `seasons:write` permission (trainer or admin)
- **Description**: Deletes a season.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | ID of the season |

**Response** `200 OK`
```json
{
  "message": "Success"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |

---

### Get Season Stats

- **Endpoint**: `GET /api/v1/seasons/{seasonId}/stats`
- **Auth**: Team member with `seasons:read` permission
- **Description**: Returns aggregated statistics for a season.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | ID of the season |

**Response** `200 OK`
```json
{
  "message": "Success",
  "stats": {
    "goalCount": 12,
    "completedGoalCount": 4,
    "openGoalCount": 5,
    "inProgressGoalCount": 3,
    "reportCount": 8,
    "memberCount": 15
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `goalCount` | integer | Total goals in the season |
| `completedGoalCount` | integer | Goals with status `completed` |
| `openGoalCount` | integer | Goals with status `open` |
| `inProgressGoalCount` | integer | Goals with status `in_progress` |
| `reportCount` | integer | Number of progress reports submitted |
| `memberCount` | integer | Number of active team members |

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `seasonId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |
