# VolleyGoals API Reference

## Overview

Base URL (local): `http://localhost:8080`
Base URL (deployed): Injected at build time via Terraform ldflags.

All authenticated endpoints require a `Authorization: Bearer <token>` header containing a valid AWS Cognito JWT.

### Common Response Envelope

All responses follow this structure:

```json
{
    "message": "success.ok | error.code",
    "...": "additional fields"
}
```

### Pagination

Endpoints that return lists use cursor-based pagination.

**Query Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `limit` | integer | Max items to return |
| `nextToken` | string | Base64-encoded cursor from a previous response |

**Response Fields:**

```json
{
    "message": "success.ok",
    "items": [...],
    "count": 10,
    "nextToken": "base64encodedcursor",
    "hasMore": true
}
```

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| `200` | OK |
| `201` | Created |
| `204` | No Content (delete success) |
| `400` | Bad Request — invalid input |
| `401` | Unauthorized — missing or invalid token |
| `403` | Forbidden — insufficient permissions |
| `404` | Not Found |
| `406` | Not Acceptable — business rule violation |
| `409` | Conflict — resource already exists |
| `503` | Service Unavailable — one or more dependencies are down (health check only) |

### Error Response

```json
{
    "message": "error.not_found",
    "error": "optional detail string"
}
```

---

## Authentication & Roles

**User Types (Cognito Groups):**

| Type | Description |
|------|-------------|
| `ADMINS` | Platform administrators — full access |
| `USERS` | Regular users — team-scoped access |

**Team Member Roles:**

| Role | Description |
|------|-------------|
| `admin` | Manages team, members, seasons, goals |
| `trainer` | Manages seasons, goals, and progress |
| `member` | Read access + own goals |

> Admins (`ADMINS` group) bypass team-role checks and can access all resources.

---

## Endpoints

---

### Health

#### `POST /health`

Check API health status by probing all service dependencies.

**Auth:** None

**Response `200`** (all dependencies healthy):
```json
{
    "message": "success.ok",
    "status": "healthy",
    "dependencies": {
        "dynamodb": "ok",
        "s3": "ok",
        "cognito": "ok"
    }
}
```

**Response `503`** (one or more dependencies failed):
```json
{
    "message": "success.ok",
    "status": "unhealthy",
    "dependencies": {
        "dynamodb": "ok",
        "s3": "error",
        "cognito": "ok"
    }
}
```

Individual dependency values: `"ok"` | `"error"`.

---

### Self (Current User)

#### `GET /api/v1/self`

Get the current authenticated user's profile and team assignments.

**Auth:** Any authenticated user

**Response `200`:**
```json
{
    "message": "success.ok",
    "user": {
        "id": "cognito-sub-uuid",
        "email": "user@example.com",
        "name": "Jane Doe",
        "picture": "https://s3.../picture.jpg",
        "preferredUsername": "janedoe",
        "enabled": true,
        "userStatus": "CONFIRMED",
        "userType": "USERS",
        "birthdate": "1990-01-15T00:00:00Z",
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-06-01T00:00:00Z"
    },
    "assignments": [
        {
            "id": "member-uuid",
            "teamId": "team-uuid",
            "role": "trainer",
            "status": "active",
            "joinedAt": "2024-03-01T00:00:00Z"
        }
    ]
}
```

---

#### `PATCH /api/v1/self`

Update the current user's profile.

**Auth:** Any authenticated user

**Request Body:**
```json
{
    "name": "Jane Doe",
    "preferredUsername": "janedoe",
    "birthdate": "1990-01-15T00:00:00Z"
}
```

All fields are optional. Only provided fields are updated.

**Response `200`:**
```json
{
    "message": "success.ok",
    "user": { ... }
}
```

---

#### `GET /api/v1/self/picture/presign`

Get a presigned S3 URL to upload the current user's profile picture.

**Auth:** Any authenticated user

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `filename` | string | Yes | Original filename |
| `contentType` | string | Yes | MIME type (e.g. `image/jpeg`) |

**Response `200`:**
```json
{
    "message": "success.ok",
    "uploadUrl": "https://s3.amazonaws.com/...presigned...",
    "key": "users/sub/profile/filename.jpg",
    "fileUrl": "https://s3.../users/sub/profile/filename.jpg"
}
```

After uploading to `uploadUrl` via `PUT`, the picture is accessible at `fileUrl`. The frontend is responsible for storing `fileUrl` back to the user profile if needed.

---

### Teams

#### `POST /api/v1/teams`

Create a new team.

**Auth:** `ADMINS` only

**Request Body:**
```json
{
    "name": "Team Alpha"
}
```

**Response `201`:**
```json
{
    "message": "success.ok",
    "team": {
        "id": "team-uuid",
        "name": "Team Alpha",
        "status": "active",
        "picture": null,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
    }
}
```

---

#### `GET /api/v1/teams`

List all teams.

**Auth:** `ADMINS` only

**Query Parameters:** Standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [ { ...team } ],
    "count": 5,
    "nextToken": null,
    "hasMore": false
}
```

---

#### `GET /api/v1/teams/:teamId`

Get a single team including its settings.

**Auth:** `ADMINS` only

**Response `200`:**
```json
{
    "message": "success.ok",
    "team": {
        "id": "team-uuid",
        "name": "Team Alpha",
        "status": "active",
        "picture": "https://s3.../team.jpg",
        "createdAt": "...",
        "updatedAt": "..."
    },
    "teamSettings": {
        "teamId": "team-uuid",
        "allowFileUploads": true,
        "allowTeamGoalComments": true,
        "allowIndividualGoalComments": false,
        "createdAt": "...",
        "updatedAt": "..."
    }
}
```

---

#### `PATCH /api/v1/teams/:teamId`

Update a team's name or status.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Request Body:**
```json
{
    "name": "New Team Name",
    "status": "inactive"
}
```

`status` values: `active` | `inactive`

**Response `200`:**
```json
{
    "message": "success.ok",
    "team": { ...updatedTeam }
}
```

---

#### `DELETE /api/v1/teams/:teamId`

Delete a team and cascade-delete all related data.

**Auth:** `ADMINS` only

> **Note:** Deletes the team, its settings, all members, all invites, all seasons, all goals, all progress reports, and all comments + comment files. S3 objects (team picture, goal pictures, comment files) are **not** deleted.

**Response `200`:**
```json
{
    "message": "success.ok"
}
```

---

#### `GET /api/v1/teams/:teamId/picture/presign`

Get a presigned S3 URL to upload the team's picture.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Query Parameters:**

| Param | Type | Required |
|-------|------|----------|
| `filename` | string | Yes |
| `contentType` | string | Yes |

**Response `200`:**
```json
{
    "message": "success.ok",
    "uploadUrl": "https://s3...",
    "key": "teams/team-uuid/picture/filename.jpg",
    "fileUrl": "https://s3.../teams/..."
}
```

---

### Team Settings

#### `PATCH /api/v1/teams/:teamId/settings`

Update team feature flags.

**Auth:** `ADMINS` only

**Request Body:**
```json
{
    "allowFileUploads": true,
    "allowTeamGoalComments": false,
    "allowIndividualGoalComments": true
}
```

All fields optional. Only provided fields are updated.

**Response `200`:**
```json
{
    "message": "success.ok",
    "teamSettings": {
        "teamId": "team-uuid",
        "allowFileUploads": true,
        "allowTeamGoalComments": false,
        "allowIndividualGoalComments": true,
        "createdAt": "...",
        "updatedAt": "..."
    }
}
```

---

### Team Members

#### `GET /api/v1/teams/:teamId/members`

List all members of a team.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Query Parameters:** Standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [
        {
            "id": "member-uuid",
            "userId": "cognito-sub",
            "teamId": "team-uuid",
            "role": "trainer",
            "status": "active",
            "joinedAt": "2024-03-01T00:00:00Z",
            "leftAt": null,
            "user": {
                "id": "cognito-sub",
                "email": "user@example.com",
                "name": "Jane Doe",
                "picture": null,
                "preferredUsername": "janedoe"
            }
        }
    ],
    "count": 3,
    "nextToken": null,
    "hasMore": false
}
```

---

#### `POST /api/v1/teams/:teamId/members`

Add a user to a team.

**Auth:** `ADMINS` only

**Request Body:**
```json
{
    "userId": "cognito-sub",
    "role": "member"
}
```

`role` values: `admin` | `trainer` | `member`

**Response `201`:**
```json
{
    "message": "success.ok",
    "teamMember": { ...teamMember }
}
```

---

#### `PATCH /api/v1/teams/:teamId/members/:memberId`

Update a member's role or status.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Request Body:**
```json
{
    "role": "trainer",
    "status": "active"
}
```

`status` values: `active` | `removed`

**Response `200`:**
```json
{
    "message": "success.ok",
    "teamMember": { ...updatedTeamMember }
}
```

---

#### `DELETE /api/v1/teams/:teamId/members/:memberId`

Remove a specific member from a team.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Response `200`:**
```json
{
    "message": "success.ok"
}
```

---

#### `DELETE /api/v1/teams/:teamId/members`

Current authenticated user leaves the team.

**Auth:** Any active team member

**Response `200`:**
```json
{
    "message": "success.ok"
}
```

**Response `406`** (if you are the only admin/trainer and cannot leave):
```json
{
    "message": "error.cannot_leave_team"
}
```

---

### Invites

#### `POST /api/v1/invites`

Create and optionally send an invitation to join a team.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Request Body:**
```json
{
    "email": "newuser@example.com",
    "teamId": "team-uuid",
    "role": "member",
    "sendEmail": true,
    "message": "We'd love to have you on the team!"
}
```

`role` values: `admin` | `trainer` | `member`
`sendEmail`: If `true`, an invitation email is sent via SES.

**Response `201`:**
```json
{
    "message": "success.ok",
    "invite": {
        "id": "invite-uuid",
        "teamId": "team-uuid",
        "email": "newuser@example.com",
        "role": "member",
        "status": "pending",
        "token": "abc123hash",
        "message": "We'd love...",
        "invitedBy": "cognito-sub",
        "expiresAt": "2024-02-01T00:00:00Z",
        "createdAt": "...",
        "updatedAt": "..."
    }
}
```

**Response `409`** if a pending invite for that email+team already exists.

---

#### `GET /api/v1/teams/:teamId/invites`

List all invites for a team.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Query Parameters:** Standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [ { ...invite } ],
    "count": 2,
    "nextToken": null,
    "hasMore": false
}
```

---

#### `GET /api/v1/invites/:inviteToken`

Get invite details by token (used on the acceptance page).

**Auth:** None (public)

**Response `200`:**
```json
{
    "message": "success.ok",
    "invite": { ...invite },
    "member": null
}
```

`member` is populated if the invited email already has a team membership.

**Response `404`** if the token is invalid or the invite has expired/been revoked.

---

#### `POST /api/v1/invites/complete`

Accept or decline an invitation.

**Auth:** None (public)

**Request Body:**
```json
{
    "token": "abc123hash",
    "email": "newuser@example.com",
    "accepted": true
}
```

**Response `200`:**
```json
{
    "message": "success.ok",
    "invite": { ...updatedInvite },
    "member": { ...teamMember },
    "userCreated": true,
    "temporaryPassword": "TempPass123!"
}
```

- `userCreated` and `temporaryPassword` are only present when a new Cognito user was created for the invited email.
- If `accepted` is `false`, the invite is marked `declined` and no member is created.

---

#### `PATCH /api/v1/invites/:inviteId`

Resend the invitation email.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Response `200`:**
```json
{
    "message": "success.ok",
    "invite": { ...invite }
}
```

---

#### `DELETE /api/v1/invites/:inviteId`

Revoke (cancel) an invitation.

**Auth:** `ADMINS` or team `admin`/`trainer`

**Response `200`:**
```json
{
    "message": "success.ok",
    "invite": { ...revokedInvite }
}
```

---

### Users (Admin)

#### `GET /api/v1/users`

List all Cognito users.

**Auth:** `ADMINS` only

**Query Parameters:**

| Param | Type | Description |
|-------|------|-------------|
| `groupName` | string | Filter by group: `admin`, `admins`, `user`, or `users` |
| `paginationToken` | string | Token from previous response |

**Response `200`:**
```json
{
    "message": "success.ok",
    "users": [ { ...user } ],
    "paginationToken": "nextpagetoken"
}
```

---

#### `GET /api/v1/users/:userSub`

Get a single user by their Cognito Sub (UUID).

**Auth:** `ADMINS` only

**Response `200`:**
```json
{
    "message": "success.ok",
    "user": {
        "id": "cognito-sub",
        "email": "user@example.com",
        "name": "Jane Doe",
        "picture": null,
        "preferredUsername": "janedoe",
        "enabled": true,
        "userStatus": "CONFIRMED",
        "userType": "USERS",
        "birthdate": null,
        "createdAt": "...",
        "updatedAt": "..."
    },
    "memberships": [
        {
            "id": "member-uuid",
            "teamId": "team-uuid",
            "role": "trainer",
            "status": "active"
        }
    ]
}
```

---

#### `PATCH /api/v1/users/:userSub`

Update a user's type or enabled state.

**Auth:** `ADMINS` only

**Request Body:**
```json
{
    "userType": "ADMINS",
    "enabled": false
}
```

`userType` values: `ADMINS` | `USERS`

> **Note:** Promoting a user to `ADMINS` automatically removes all their team memberships.

**Response `200`:**
```json
{
    "message": "success.ok",
    "user": { ...updatedUser }
}
```

---

#### `DELETE /api/v1/users/:userSub`

Delete a Cognito user and remove all their team memberships.

**Auth:** `ADMINS` only

**Response `200`:**
```json
{
    "message": "success.ok"
}
```

---

### Seasons

#### `POST /api/v1/seasons`

Create a new season for a team.

**Auth:** Team `admin` or `trainer`

**Request Body:**
```json
{
    "teamId": "team-uuid",
    "name": "Spring 2024",
    "startDate": "2024-03-01T00:00:00Z",
    "endDate": "2024-06-30T00:00:00Z"
}
```

**Response `201`:**
```json
{
    "message": "success.ok",
    "season": {
        "id": "season-uuid",
        "teamId": "team-uuid",
        "name": "Spring 2024",
        "startDate": "2024-03-01T00:00:00Z",
        "endDate": "2024-06-30T00:00:00Z",
        "status": "planned",
        "createdAt": "...",
        "updatedAt": "..."
    }
}
```

---

#### `GET /api/v1/seasons`

List all seasons for a team.

**Auth:** Any active team member

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `teamId` | string | Yes | Filter by team |

Plus standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [ { ...season } ],
    "count": 3,
    "nextToken": null,
    "hasMore": false
}
```

---

#### `GET /api/v1/seasons/:seasonId`

Get a single season.

**Auth:** Any active team member

**Response `200`:**
```json
{
    "message": "success.ok",
    "season": { ...season }
}
```

---

#### `PATCH /api/v1/seasons/:seasonId`

Update a season.

**Auth:** Team `admin` or `trainer`

**Request Body:**
```json
{
    "name": "Updated Name",
    "startDate": "2024-03-01T00:00:00Z",
    "endDate": "2024-07-31T00:00:00Z",
    "status": "active"
}
```

`status` values: `planned` | `active` | `completed` | `archived`

All fields optional.

**Response `200`:**
```json
{
    "message": "success.ok",
    "season": { ...updatedSeason }
}
```

---

#### `DELETE /api/v1/seasons/:seasonId`

Delete a season.

**Auth:** Team `admin` or `trainer`

**Response `200`:**
```json
{
    "message": "success.ok"
}
```

---

### Goals

#### `POST /api/v1/seasons/:seasonId/goals`

Create a goal within a season.

**Auth:**
- `team` goals: team `admin` or `trainer`
- `individual` goals: any team member

**Request Body:**
```json
{
    "type": "individual",
    "title": "Improve my serve",
    "description": "Reach 80% first-serve accuracy by end of season."
}
```

`type` values: `individual` | `team`

**Response `201`:**
```json
{
    "message": "success.ok",
    "goal": {
        "id": "goal-uuid",
        "seasonId": "season-uuid",
        "ownerId": "cognito-sub",
        "goalType": "individual",
        "title": "Improve my serve",
        "description": "Reach 80% first-serve accuracy by end of season.",
        "status": "open",
        "picture": null,
        "createdBy": "cognito-sub",
        "createdAt": "...",
        "updatedAt": "..."
    }
}
```

---

#### `GET /api/v1/seasons/:seasonId/goals`

List goals for a season.

**Auth:** Any active team member

**Query Parameters:** Standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [ { ...goal } ],
    "count": 8,
    "nextToken": null,
    "hasMore": false
}
```

---

#### `GET /api/v1/seasons/:seasonId/goals/:goalId`

Get a single goal.

**Auth:** Any active team member

**Response `200`:**
```json
{
    "message": "success.ok",
    "goal": { ...goal }
}
```

---

#### `PATCH /api/v1/seasons/:seasonId/goals/:goalId`

Update a goal.

**Auth:** Goal owner or team `admin`/`trainer`

**Request Body:**
```json
{
    "ownerId": "other-cognito-sub",
    "title": "Updated title",
    "description": "Updated description",
    "status": "in_progress"
}
```

`status` values: `open` | `in_progress` | `completed` | `archived`

All fields optional.

**Response `200`:**
```json
{
    "message": "success.ok",
    "goal": { ...updatedGoal }
}
```

---

#### `DELETE /api/v1/seasons/:seasonId/goals/:goalId`

Delete a goal.

**Auth:** Goal owner or team `admin`/`trainer`

**Response `204`:** Empty body.

---

#### `GET /api/v1/seasons/:seasonId/goals/:goalId/picture/presign`

Get a presigned S3 URL to upload a goal picture.

**Auth:** Goal owner or team `admin`/`trainer`

**Query Parameters:**

| Param | Type | Required |
|-------|------|----------|
| `filename` | string | Yes |
| `contentType` | string | Yes |

**Response `200`:**
```json
{
    "message": "success.ok",
    "uploadUrl": "https://s3...",
    "key": "goals/goal-uuid/picture/filename.jpg",
    "fileUrl": "https://s3.../goals/..."
}
```

---

### Progress Reports

#### `POST /api/v1/seasons/:seasonId/progress-reports`

Create a progress report for a season.

**Auth:** Any active team member

**Request Body:**
```json
{
    "summary": "Good week overall",
    "details": "Worked on serve accuracy and blocking drills.",
    "progress": [
        { "goalId": "goal-uuid-1", "rating": 4 },
        { "goalId": "goal-uuid-2", "rating": 2 }
    ]
}
```

`progress` is optional. Each entry rates a specific goal (1–5 scale by convention).

**Response `201`:**
```json
{
    "message": "success.ok",
    "progressReport": {
        "id": "report-uuid",
        "seasonId": "season-uuid",
        "authorId": "cognito-sub",
        "summary": "Good week overall",
        "details": "Worked on serve accuracy and blocking drills.",
        "createdAt": "2024-04-01T00:00:00Z",
        "updatedAt": "2024-04-01T00:00:00Z"
    }
}
```

---

#### `GET /api/v1/seasons/:seasonId/progress-reports`

List progress reports for a season.

**Auth:** Any active team member

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `authorId` | string | No | Filter by author (Cognito Sub) |
| `summary` | string | No | Partial match on summary text |

Plus standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [ { ...progressReport } ],
    "count": 4,
    "nextToken": null,
    "hasMore": false
}
```

---

#### `GET /api/v1/seasons/:seasonId/progress-reports/:reportId`

Get a single progress report.

**Auth:** Any active team member

**Response `200`:**
```json
{
    "message": "success.ok",
    "progressReport": { ...progressReport }
}
```

**Response `404`** (`error.progressReport.notFound`) if the report does not exist or belongs to a different season.

---

#### `PATCH /api/v1/seasons/:seasonId/progress-reports/:reportId`

Update a progress report.

**Auth:** Report author or team `admin`/`trainer`

**Request Body:**
```json
{
    "summary": "Updated summary",
    "details": "Updated details.",
    "progress": [
        { "goalId": "goal-uuid-1", "rating": 5 }
    ]
}
```

All fields optional. If `progress` is provided, the existing progress entries for this report are replaced entirely.

**Response `200`:**
```json
{
    "message": "success.ok",
    "progressReport": { ...updatedProgressReport }
}
```

**Response `403`** if the requester is neither the author nor an admin/trainer.
**Response `404`** (`error.progressReport.notFound`) if not found.

---

#### `DELETE /api/v1/seasons/:seasonId/progress-reports/:reportId`

Delete a progress report and its associated progress entries.

**Auth:** Report author or team `admin`/`trainer`

**Response `204`:** Empty body.

**Response `403`** if the requester is neither the author nor an admin/trainer.
**Response `404`** (`error.progressReport.notFound`) if not found.

---

### Comments

#### `POST /api/v1/comments`

Create a comment on a goal or progress report.

**Auth:** Any active team member

**Request Body:**
```json
{
    "commentType": "Goal",
    "targetId": "goal-uuid",
    "content": "Great progress on this goal!"
}
```

`commentType` values: `Goal` | `ProgressReport`

**Rules:**
- For `Goal` targets: the requester must be a member of the goal's team. Additionally:
  - If the goal is a `team` goal: `TeamSettings.allowTeamGoalComments` must be `true`, otherwise `403`.
  - If the goal is an `individual` goal: `TeamSettings.allowIndividualGoalComments` must be `true`, otherwise `403`.
- For `ProgressReport` targets: no TeamSettings gate — any team member may comment.

**Response `201`:**
```json
{
    "message": "success.ok",
    "comment": {
        "id": "comment-uuid",
        "authorId": "cognito-sub",
        "commentType": "Goal",
        "targetId": "goal-uuid",
        "content": "Great progress on this goal!",
        "createdAt": "2024-04-01T12:00:00Z",
        "updatedAt": "2024-04-01T12:00:00Z"
    }
}
```

**Response `403`** (`error.comment.disabled`) if team settings disallow comments for the target goal type.
**Response `404`** if the target does not exist.

---

#### `GET /api/v1/comments`

List comments for a specific target.

**Auth:** Any active team member of the target's team

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `targetId` | string | **Yes** | ID of the goal or progress report |
| `commentType` | string | **Yes** | `Goal` or `ProgressReport` |
| `authorId` | string | No | Filter by author |

Plus standard pagination params.

**Response `200`:**
```json
{
    "message": "success.ok",
    "items": [ { ...comment } ],
    "count": 3,
    "nextToken": null,
    "hasMore": false
}
```

**Response `400`** if `targetId` or `commentType` are missing.

---

#### `GET /api/v1/comments/:commentId`

Get a single comment.

**Auth:** Any active team member of the target's team

**Response `200`:**
```json
{
    "message": "success.ok",
    "comment": { ...comment }
}
```

**Response `404`** (`error.comment.notFound`) if the comment does not exist.

---

#### `PATCH /api/v1/comments/:commentId`

Update a comment's content.

**Auth:** Comment author or team `admin`/`trainer`

**Request Body:**
```json
{
    "content": "Updated comment text."
}
```

**Response `200`:**
```json
{
    "message": "success.ok",
    "comment": { ...updatedComment }
}
```

**Response `403`** if the requester is neither the author nor an admin/trainer.
**Response `404`** (`error.comment.notFound`) if not found.

---

#### `DELETE /api/v1/comments/:commentId`

Delete a comment.

**Auth:** Comment author or team `admin`/`trainer`

**Response `204`:** Empty body.

**Response `403`** if the requester is neither the author nor an admin/trainer.
**Response `404`** (`error.comment.notFound`) if not found.

---

#### `GET /api/v1/comments/:commentId/file/presign`

Get a presigned S3 URL to upload a file attachment to a comment. Also creates a `CommentFile` record in the database.

**Auth:** Comment author or team `admin`/`trainer`

**Query Parameters:**

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `filename` | string | Yes | Original filename (used for extension) |
| `contentType` | string | Yes | MIME type (e.g. `image/png`, `application/pdf`) |

**Response `200`:**
```json
{
    "message": "success.ok",
    "uploadUrl": "https://s3.amazonaws.com/...presigned...",
    "key": "comments/comment-uuid/generatedfilename.png",
    "commentFile": {
        "id": "comment-file-uuid",
        "commentId": "comment-uuid",
        "storageKey": "comments/comment-uuid/generatedfilename.png",
        "createdAt": "2024-04-01T12:00:00Z"
    }
}
```

**Response `403`** if the requester is neither the author nor an admin/trainer.
**Response `404`** (`error.comment.notFound`) if the comment does not exist.

---

## Data Models

### User

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | Cognito Sub (UUID) |
| `email` | string | |
| `name` | string \| null | |
| `picture` | string \| null | S3 URL |
| `preferredUsername` | string \| null | |
| `enabled` | boolean | |
| `userStatus` | string | Cognito status: `CONFIRMED`, `FORCE_CHANGE_PASSWORD`, etc. |
| `userType` | string | `ADMINS` \| `USERS` |
| `birthdate` | string \| null | ISO 8601 |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### Team

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `name` | string | |
| `status` | string | `active` \| `inactive` |
| `picture` | string \| null | S3 URL |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### TeamMember

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `userId` | string | Cognito Sub |
| `teamId` | string | UUID |
| `role` | string | `admin` \| `trainer` \| `member` |
| `status` | string | `active` \| `invited` \| `removed` \| `left` |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |
| `joinedAt` | string \| null | ISO 8601 |
| `leftAt` | string \| null | ISO 8601 |

### TeamSettings

| Field | Type | Notes |
|-------|------|-------|
| `teamId` | string | UUID |
| `allowFileUploads` | boolean | |
| `allowTeamGoalComments` | boolean | |
| `allowIndividualGoalComments` | boolean | |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### Invite

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `teamId` | string | UUID |
| `email` | string | |
| `role` | string | `admin` \| `trainer` \| `member` |
| `status` | string | `pending` \| `accepted` \| `declined` \| `revoked` \| `expired` |
| `token` | string | Hash-based token for invite URL |
| `message` | string \| null | Optional personal message |
| `invitedBy` | string | Cognito Sub |
| `acceptedBy` | string \| null | Cognito Sub |
| `revokedBy` | string \| null | Cognito Sub |
| `expiresAt` | string | ISO 8601 |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |
| `acceptedAt` | string \| null | ISO 8601 |
| `declinedAt` | string \| null | ISO 8601 |
| `revokedAt` | string \| null | ISO 8601 |

### Season

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `teamId` | string | UUID |
| `name` | string | |
| `startDate` | string | ISO 8601 |
| `endDate` | string | ISO 8601 |
| `status` | string | `planned` \| `active` \| `completed` \| `archived` |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### Goal

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `seasonId` | string | UUID |
| `ownerId` | string | Cognito Sub |
| `goalType` | string | `individual` \| `team` |
| `title` | string | |
| `description` | string | |
| `status` | string | `open` \| `in_progress` \| `completed` \| `archived` |
| `picture` | string \| null | S3 URL |
| `createdBy` | string | Cognito Sub |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### ProgressReport

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `seasonId` | string | UUID |
| `authorId` | string | Cognito Sub |
| `summary` | string | |
| `details` | string | |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### Progress

Represents a goal rating within a progress report. Written on create/update of a progress report via the `progress` array; not returned as a top-level resource.

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `progressReportId` | string | UUID |
| `goalId` | string | UUID |
| `rating` | integer | Goal rating (int8) |

### Comment

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `authorId` | string | Cognito Sub |
| `commentType` | string | `Goal` \| `ProgressReport` |
| `targetId` | string | UUID of the goal or progress report |
| `content` | string | |
| `createdAt` | string | ISO 8601 |
| `updatedAt` | string | ISO 8601 |

### CommentFile

| Field | Type | Notes |
|-------|------|-------|
| `id` | string | UUID |
| `commentId` | string | UUID |
| `storageKey` | string | S3 key (`comments/{commentId}/{filename}`) |
| `createdAt` | string | ISO 8601 |

---

## File Uploads

File uploads use a two-step presign flow:

1. Call the relevant `GET .../presign` endpoint to receive a `uploadUrl` and `fileUrl`.
2. Upload the file directly to `uploadUrl` using an HTTP `PUT` request with the correct `Content-Type` header.
3. The public `fileUrl` can then be used in the frontend to display the file.

The backend does **not** store the `fileUrl` automatically after upload — the frontend must persist it back via the relevant PATCH endpoint where applicable (e.g., `PATCH /api/v1/self` with a `picture` field — check implementation status).

File upload availability per team is controlled by `TeamSettings.allowFileUploads`.

---

## Environment & Deployment

The API runs in two modes selected at build time via Go ldflags:

| Mode | Framework | Auth |
|------|-----------|------|
| Local | Gin + CORS middleware | Local JWT validation via Cognito JWKS |
| Lambda | AWS Lambda Go | API Gateway Authorizer (token pre-validated) |

Tracing is provided by OpenTelemetry with X-Ray-compatible propagation (`_X_AMZN_TRACE_ID`).
