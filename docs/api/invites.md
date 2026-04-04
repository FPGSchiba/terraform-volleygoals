# Invites API

The invites system manages team membership invitations via email tokens. Some endpoints are **public** (no auth required) to allow recipients to accept invitations before they have an account.

---

### Create Invite

- **Endpoint**: `POST /api/v1/invites`
- **Auth**: Global Admin, Team Admin, or team member with `invites:write` permission. Only Global Admins can invite with `admin` role.
- **Description**: Creates a new invite for a user by email. Optionally sends an invitation email.

**Request Body**
```json
{
  "email": "newplayer@example.com",
  "teamId": "team-uuid",
  "role": "member",
  "sendEmail": true,
  "message": "We'd love to have you on the team!"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `email` | string | Yes | Email address of the invitee |
| `teamId` | string (UUID) | Yes | Team to invite the user to |
| `role` | string | Yes | Role to assign: `"admin"`, `"trainer"`, or `"member"` |
| `sendEmail` | boolean | Yes | Whether to send an invitation email via SES |
| `message` | string | No | Optional personal message included in the email |

**Response** `200 OK`
```json
{
  "message": "Success",
  "invite": {
    "id": "invite-uuid",
    "teamId": "team-uuid",
    "email": "newplayer@example.com",
    "role": "member",
    "status": "pending",
    "token": "hashed-token-string",
    "message": "We'd love to have you on the team!",
    "invitedBy": "cognito-sub-of-inviter",
    "acceptedBy": null,
    "revokedBy": null,
    "expiresAt": "2024-11-01T10:00:00Z",
    "createdAt": "2024-10-25T10:00:00Z",
    "updatedAt": "2024-10-25T10:00:00Z",
    "acceptedAt": null,
    "declinedAt": null,
    "revokedAt": null
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body |
| `403 Forbidden` | Not authorized (e.g., trainer trying to invite admin) |
| `409 Conflict` | Invite already exists, or user is already a member |

---

### Complete Invite

- **Endpoint**: `POST /api/v1/invites/complete`
- **Auth**: **Public (no auth required)**
- **Description**: Accepts or declines an invitation using the invite token. If the email does not yet have a Cognito account, one is created automatically and a temporary password is returned.

**Request Body**
```json
{
  "token": "hashed-token-string",
  "email": "newplayer@example.com",
  "accepted": true
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `token` | string | Yes | The invite token from the invitation email |
| `email` | string | Yes | The invitee's email address |
| `accepted` | boolean | Yes | `true` to accept, `false` to decline |

**Response** `200 OK`
```json
{
  "message": "Success",
  "invite": {
    "id": "invite-uuid",
    "status": "accepted",
    "acceptedBy": "cognito-sub",
    "acceptedAt": "2024-10-26T08:00:00Z"
  },
  "member": {
    "id": "member-uuid",
    "userId": "cognito-sub",
    "teamId": "team-uuid",
    "role": "member",
    "status": "active",
    "createdAt": "2024-10-26T08:00:00Z",
    "updatedAt": "2024-10-26T08:00:00Z",
    "joinedAt": "2024-10-26T08:00:00Z",
    "leftAt": null
  },
  "userCreated": true,
  "temporaryPassword": "Temp@12345"
}
```

> **Note**: `temporaryPassword` is only present if a new Cognito user was created. The user must change this password on first login.

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid token, expired invite, or invite already completed |
| `404 Not Found` | Token not found |

---

### Get Invite by Token

- **Endpoint**: `GET /api/v1/invites/{token}`
- **Auth**: **Public (no auth required)** — `token` is the invite **token** string, not a UUID
- **Description**: Looks up an invite by its token. Returns the invite and team membership if already accepted. Used to validate a token before completing.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `token` | string | Yes | The invite token (from the invitation URL) |

**Response** `200 OK` (if already accepted)
```json
{
  "message": "Success",
  "invite": {
    "id": "invite-uuid",
    "status": "accepted",
    "email": "newplayer@example.com",
    "teamId": "team-uuid",
    "role": "member",
    "expiresAt": "2024-11-01T10:00:00Z"
  },
  "member": {
    "id": "member-uuid",
    "userId": "cognito-sub",
    "teamId": "team-uuid",
    "role": "member",
    "status": "active"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Token is expired or invalid |
| `404 Not Found` | Token not found or invite not yet accepted |

---

### Get Team Invites

- **Endpoint**: `GET /api/v1/teams/{teamId}/invites`
- **Auth**: Global Admin or team member with `invites:write` permission
- **Description**: Returns a paginated list of all invites for a team.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `teamId` | string (UUID) | Yes | Team whose invites to list |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `status` | string | No | Filter by status: `pending`, `accepted`, `declined`, `revoked`, `expired` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "invite-uuid",
      "teamId": "team-uuid",
      "email": "newplayer@example.com",
      "role": "member",
      "status": "pending",
      "token": "hashed-token-string",
      "invitedBy": "cognito-sub",
      "expiresAt": "2024-11-01T10:00:00Z",
      "createdAt": "2024-10-25T10:00:00Z",
      "updatedAt": "2024-10-25T10:00:00Z"
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
| `403 Forbidden` | Not authorized |

---

### Revoke Invite

- **Endpoint**: `DELETE /api/v1/invites/{inviteId}`
- **Auth**: Global Admin or team member with `invites:write` permission
- **Description**: Revokes a pending invite so it can no longer be accepted.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `inviteId` | string (UUID) | Yes | ID of the invite |

**Response** `200 OK`
```json
{
  "message": "Success",
  "invite": {
    "id": "invite-uuid",
    "status": "revoked",
    "revokedBy": "cognito-sub",
    "revokedAt": "2024-10-27T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `inviteId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Invite not found |

---

### Resend Invite

- **Endpoint**: `PATCH /api/v1/invites/{inviteId}`
- **Auth**: Global Admin or team member with `invites:write` permission
- **Description**: Resends the invitation email for a pending invite.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `inviteId` | string (UUID) | Yes | ID of the invite |

**Response** `200 OK`
```json
{
  "message": "Success",
  "invite": {
    "id": "invite-uuid",
    "status": "pending",
    "email": "newplayer@example.com"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `inviteId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Invite not found |

---

## Invite Status Values

| Status | Description |
|--------|-------------|
| `pending` | Sent but not yet acted on |
| `accepted` | Invitee accepted and is now a team member |
| `declined` | Invitee declined the invitation |
| `revoked` | Admin revoked the invite before it was completed |
| `expired` | Invite passed its expiration date |
