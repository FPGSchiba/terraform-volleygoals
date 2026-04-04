# Goals API

Goals are nested under seasons: `/api/v1/seasons/{seasonId}/goals`.

A goal can be either **individual** (owned by one player) or **team** (shared). Access is controlled by the goal owner and the caller's role.

---

### Create Goal

- **Endpoint**: `POST /api/v1/seasons/{seasonId}/goals`
- **Auth**: Team member with `goals:write` permission
- **Description**: Creates a new goal within a season. The caller becomes the owner by default; a different `ownerId` can be specified if the caller has write permission.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the goal belongs to |

**Request Body**
```json
{
  "type": "individual",
  "title": "Improve serve accuracy",
  "description": "Achieve 80% first-serve accuracy by end of season",
  "ownerId": "cognito-sub"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | `"individual"` or `"team"` |
| `title` | string | Yes | Short goal title |
| `description` | string | Yes | Detailed goal description |
| `ownerId` | string | No | Cognito sub of the owner (defaults to caller) |

**Response** `201 Created`
```json
{
  "message": "Success",
  "goal": {
    "id": "goal-uuid",
    "seasonId": "season-uuid",
    "ownerId": "cognito-sub",
    "goalType": "individual",
    "picture": "",
    "title": "Improve serve accuracy",
    "description": "Achieve 80% first-serve accuracy by end of season",
    "status": "open",
    "createdBy": "cognito-sub",
    "createdAt": "2024-09-10T08:00:00Z",
    "updatedAt": "2024-09-10T08:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or `seasonId` |
| `403 Forbidden` | Not authorized |

---

### Get Goal

- **Endpoint**: `GET /api/v1/seasons/{seasonId}/goals/{goalId}`
- **Auth**: Global Admin, or team member with `goals:read` permission for own goals (ownership-aware)
- **Description**: Returns a single goal. Non-admins can only read goals they own or have explicit access to via ownership policy.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the goal belongs to |
| `goalId` | string (UUID) | Yes | ID of the goal |

**Response** `200 OK`
```json
{
  "message": "Success",
  "goal": {
    "id": "goal-uuid",
    "seasonId": "season-uuid",
    "ownerId": "cognito-sub",
    "goalType": "individual",
    "picture": "https://cdn.example.com/goals/goal-uuid/picture.jpg",
    "title": "Improve serve accuracy",
    "description": "Achieve 80% first-serve accuracy by end of season",
    "status": "in_progress",
    "createdBy": "cognito-sub",
    "createdAt": "2024-09-10T08:00:00Z",
    "updatedAt": "2024-10-01T09:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `seasonId` or `goalId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Goal not found in the given season |

---

### List Goals

- **Endpoint**: `GET /api/v1/seasons/{seasonId}/goals`
- **Auth**: Team member with `goals:read` permission. Non-admins can only see their own goals.
- **Description**: Returns a paginated list of goals for a season. Non-admin callers are automatically filtered to their own goals.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season to list goals for |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `status` | string | No | Filter by status: `open`, `in_progress`, `completed`, `archived` |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "goal-uuid",
      "seasonId": "season-uuid",
      "ownerId": "cognito-sub",
      "goalType": "individual",
      "picture": "",
      "title": "Improve serve accuracy",
      "description": "Achieve 80% first-serve accuracy by end of season",
      "status": "in_progress",
      "createdBy": "cognito-sub",
      "createdAt": "2024-09-10T08:00:00Z",
      "updatedAt": "2024-10-01T09:00:00Z",
      "owner": {
        "id": "cognito-sub",
        "name": "Jane Doe",
        "preferredUsername": "jane_v",
        "picture": "https://cdn.example.com/users/jane.jpg"
      },
      "completionPercentage": 60
    }
  ],
  "count": 1,
  "nextToken": "",
  "hasMore": false
}
```

| Field | Type | Description |
|-------|------|-------------|
| `owner` | object | Enriched user data (name, picture) for the goal owner |
| `completionPercentage` | integer | Average completion rating (0–100) derived from progress entries |

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `seasonId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |

---

### Update Goal

- **Endpoint**: `PATCH /api/v1/seasons/{seasonId}/goals/{goalId}`
- **Auth**: Global Admin, or team member with `goals:write` on their own goals
- **Description**: Updates a goal's owner, title, description, or status. Status changes emit an activity event.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the goal belongs to |
| `goalId` | string (UUID) | Yes | ID of the goal |

**Request Body** (all fields optional)
```json
{
  "ownerId": "new-owner-cognito-sub",
  "title": "Updated title",
  "description": "Updated description",
  "status": "completed"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ownerId` | string | No | Reassign to a different owner |
| `title` | string | No | New title |
| `description` | string | No | New description |
| `status` | string | No | `"open"`, `"in_progress"`, `"completed"`, or `"archived"` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "goal": {
    "id": "goal-uuid",
    "seasonId": "season-uuid",
    "ownerId": "cognito-sub",
    "goalType": "individual",
    "picture": "",
    "title": "Updated title",
    "description": "Updated description",
    "status": "completed",
    "createdBy": "cognito-sub",
    "createdAt": "2024-09-10T08:00:00Z",
    "updatedAt": "2024-11-01T12:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing path params |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Goal not found in the given season |

---

### Delete Goal

- **Endpoint**: `DELETE /api/v1/seasons/{seasonId}/goals/{goalId}`
- **Auth**: Global Admin, or team member with `goals:delete` permission for own goals
- **Description**: Deletes a goal.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the goal belongs to |
| `goalId` | string (UUID) | Yes | ID of the goal |

**Response** `204 No Content`

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing path params |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Goal not found in the given season |

---

### Upload Goal Picture (Presign)

- **Endpoint**: `GET /api/v1/seasons/{seasonId}/goals/{goalId}/picture/presign`
- **Auth**: Global Admin, or team member with `goals:write` on their own goal
- **Description**: Generates a presigned S3 upload URL for a goal picture. After uploading to S3 via the returned URL, the goal's `picture` field is automatically updated.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the goal belongs to |
| `goalId` | string (UUID) | Yes | ID of the goal |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filename` | string | Yes | Original filename (e.g., `goal-pic.jpg`) |
| `contentType` | string | Yes | MIME type (e.g., `image/jpeg`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "uploadUrl": "https://s3.amazonaws.com/bucket/goals/goal-uuid/goal-pic.jpg?X-Amz-...",
  "key": "goals/goal-uuid/goal-pic.jpg",
  "fileUrl": "https://cdn.example.com/goals/goal-uuid/goal-pic.jpg"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing path params, `filename`, or `contentType` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Goal not found in the given season |

---

## Goal Status Values

| Status | Description |
|--------|-------------|
| `open` | Goal created, not yet started |
| `in_progress` | Actively being worked on |
| `completed` | Goal achieved |
| `archived` | Closed without completion |

## Goal Type Values

| Type | Description |
|------|-------------|
| `individual` | Personal goal for one team member |
| `team` | Shared goal for the whole team |
