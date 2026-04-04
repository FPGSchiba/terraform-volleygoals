# Teams API

Covers Teams, Team Settings, and Team Members.

---

## Teams

### List Teams

- **Endpoint**: `GET /api/v1/teams`
- **Auth**: Global Admin only
- **Description**: Returns a paginated list of all teams.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Number of items per page (default varies) |
| `nextToken` | string | No | Cursor from a previous response for pagination |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "team-uuid",
      "name": "Volleyball Team A",
      "status": "active",
      "picture": "https://cdn.example.com/teams/team-uuid/logo.jpg",
      "tenantId": "tenant-uuid",
      "createdAt": "2024-01-15T10:00:00Z",
      "updatedAt": "2024-06-01T12:00:00Z",
      "deletedAt": null
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
| `403 Forbidden` | Caller is not a global admin |

---

### Get Team

- **Endpoint**: `GET /api/v1/teams/{teamId}`
- **Auth**: Global Admin or team member with `teams:read` permission
- **Description**: Returns a single team and its settings.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Response** `200 OK`
```json
{
  "message": "Success",
  "team": {
    "id": "team-uuid",
    "name": "Volleyball Team A",
    "status": "active",
    "picture": "https://cdn.example.com/teams/team-uuid/logo.jpg",
    "tenantId": "tenant-uuid",
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-06-01T12:00:00Z",
    "deletedAt": null
  },
  "teamSettings": {
    "id": "settings-uuid",
    "teamId": "team-uuid",
    "allowFileUploads": true,
    "allowTeamGoalComments": true,
    "allowIndividualGoalComments": false,
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-06-01T12:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Team or team settings not found |

---

### Create Team

- **Endpoint**: `POST /api/v1/teams`
- **Auth**: Global Admin only
- **Description**: Creates a new team with default settings.

**Request Body**
```json
{
  "name": "Volleyball Team A"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique team name |

**Response** `200 OK`
```json
{
  "message": "Team created successfully",
  "team": {
    "id": "team-uuid",
    "name": "Volleyball Team A",
    "status": "active",
    "picture": "",
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-01-15T10:00:00Z",
    "deletedAt": null
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or team name already exists |
| `403 Forbidden` | Not a global admin |

---

### Update Team

- **Endpoint**: `PATCH /api/v1/teams/{teamId}`
- **Auth**: Global Admin only
- **Description**: Updates a team's name or status.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Request Body** (all fields optional)
```json
{
  "name": "New Team Name",
  "status": "inactive"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | New team name |
| `status` | string | No | `"active"` or `"inactive"` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "team": {
    "id": "team-uuid",
    "name": "New Team Name",
    "status": "inactive",
    "picture": "",
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-06-01T12:00:00Z",
    "deletedAt": null
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `teamId` |
| `403 Forbidden` | Not a global admin |
| `404 Not Found` | Team not found |

---

### Delete Team

- **Endpoint**: `DELETE /api/v1/teams/{teamId}`
- **Auth**: Global Admin only
- **Description**: Deletes a team and all associated data.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Response** `200 OK`
```json
{
  "message": "Team deleted successfully"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId` |
| `403 Forbidden` | Not a global admin |

---

### Upload Team Picture (Presign)

- **Endpoint**: `GET /api/v1/teams/{teamId}/picture/presign`
- **Auth**: Global Admin or team member with `teams:write` permission
- **Description**: Generates a presigned S3 upload URL for the team's picture. After uploading directly to S3 using the returned URL, the team's `picture` field is automatically updated.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filename` | string | Yes | Original filename (e.g., `logo.png`) |
| `contentType` | string | Yes | MIME type (e.g., `image/png`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "uploadUrl": "https://s3.amazonaws.com/bucket/teams/team-uuid/logo.png?X-Amz-...",
  "key": "teams/team-uuid/logo.png",
  "fileUrl": "https://cdn.example.com/teams/team-uuid/logo.png"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId`, `filename`, or `contentType` |
| `403 Forbidden` | Not authorized |

---

## Team Settings

### Update Team Settings

- **Endpoint**: `PATCH /api/v1/teams/{teamId}/settings`
- **Auth**: Global Admin or team member with `team_settings:write` permission
- **Description**: Updates configurable settings for a team.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Request Body** (all fields optional)
```json
{
  "allowFileUploads": true,
  "allowTeamGoalComments": true,
  "allowIndividualGoalComments": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `allowFileUploads` | boolean | No | Whether members can upload files to comments |
| `allowTeamGoalComments` | boolean | No | Whether comments are enabled on team goals |
| `allowIndividualGoalComments` | boolean | No | Whether comments are enabled on individual goals |

**Response** `200 OK`
```json
{
  "message": "Success",
  "teamSettings": {
    "id": "settings-uuid",
    "teamId": "team-uuid",
    "allowFileUploads": true,
    "allowTeamGoalComments": true,
    "allowIndividualGoalComments": false,
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": "2024-06-01T12:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `teamId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Team settings not found |

---

## Team Members

### List Team Members

- **Endpoint**: `GET /api/v1/teams/{teamId}/members`
- **Auth**: Global Admin or team member with `members:read` permission
- **Description**: Returns a paginated list of team members enriched with user profile data. Regular members (`role=member`) receive a reduced public view.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `nameContains` | string | No | Filter by user name (case-insensitive) |
| `emailContains` | string | No | Filter by email (case-insensitive) |

**Response** `200 OK` (admin/trainer view)
```json
{
  "message": "Success",
  "items": [
    {
      "id": "member-uuid",
      "userId": "cognito-sub",
      "role": "trainer",
      "status": "active",
      "userStatus": "CONFIRMED",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "picture": "https://cdn.example.com/users/jane.jpg",
      "preferredUsername": "jane_v",
      "birthdate": "1990-05-20T00:00:00Z",
      "joinedAt": "2024-02-01T09:00:00Z"
    }
  ],
  "count": 1,
  "nextToken": "",
  "hasMore": false
}
```

> **Note**: When the caller has role `member`, the response items only include `id`, `userId`, `name`, `preferredUsername`, `picture`, `email`.

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId` |
| `403 Forbidden` | Not authorized |

---

### Add Team Member

- **Endpoint**: `POST /api/v1/teams/{teamId}/members`
- **Auth**: Global Admin or team member with `members:write` permission
- **Description**: Directly adds an existing user to a team by user ID (no invite flow).

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Request Body**
```json
{
  "userId": "cognito-sub",
  "role": "member"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userId` | string | Yes | Cognito sub of the user to add |
| `role` | string | Yes | `"admin"`, `"trainer"`, or `"member"` |

**Response** `201 Created`
```json
{
  "message": "Success",
  "teamMember": {
    "id": "member-uuid",
    "userId": "cognito-sub",
    "teamId": "team-uuid",
    "role": "member",
    "status": "active",
    "createdAt": "2024-06-01T10:00:00Z",
    "updatedAt": "2024-06-01T10:00:00Z",
    "joinedAt": "2024-06-01T10:00:00Z",
    "leftAt": null
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | User not found |

---

### Update Team Member

- **Endpoint**: `PATCH /api/v1/teams/{teamId}/members/{memberId}`
- **Auth**: Global Admin or team member with `members:write` permission. Assigning `admin` role requires `members:delete` permission.
- **Description**: Updates a team member's role or status.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |
| `memberId` | string (UUID) | Yes | ID of the team member record |

**Request Body** (all fields optional)
```json
{
  "role": "trainer",
  "status": "active"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `role` | string | No | `"admin"`, `"trainer"`, or `"member"` |
| `status` | string | No | `"active"`, `"invited"`, `"removed"`, or `"left"` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "teamMember": {
    "id": "member-uuid",
    "userId": "cognito-sub",
    "teamId": "team-uuid",
    "role": "trainer",
    "status": "active",
    "createdAt": "2024-06-01T10:00:00Z",
    "updatedAt": "2024-06-10T08:00:00Z",
    "joinedAt": "2024-06-01T10:00:00Z",
    "leftAt": null
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing path params |
| `403 Forbidden` | Not authorized |

---

### Remove Team Member

- **Endpoint**: `DELETE /api/v1/teams/{teamId}/members/{memberId}`
- **Auth**: Global Admin or team member with `members:write` permission
- **Description**: Removes a member from the team.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |
| `memberId` | string (UUID) | Yes | ID of the team member record |

**Response** `200 OK`
```json
{
  "message": "Success"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing path params |
| `403 Forbidden` | Not authorized |

---

### Leave Team

- **Endpoint**: `DELETE /api/v1/teams/{teamId}/members/leave`
- **Auth**: Any authenticated team member
- **Description**: The currently authenticated user leaves the team. Admins and Trainers can only leave if another Admin or Trainer exists in the team.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | ID of the team |

**Response** `200 OK`
```json
{
  "message": "Success"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `teamId` |
| `401 Unauthorized` | Not authenticated |
| `403 Forbidden` | Not a member of this team |
| `406 Not Acceptable` | Last admin/trainer cannot leave |

---

### Get Team Activity

- **Endpoint**: `GET /api/v1/teams/{teamId}/activity`
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

