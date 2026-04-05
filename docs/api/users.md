# Users API

The Users API provides admin-only access to manage Cognito user accounts. All endpoints require **Global Admin** access.

---

### List Users

- **Endpoint**: `GET /api/v1/users`
- **Auth**: Global Admin only
- **Description**: Returns a paginated list of Cognito users. Can be filtered to a specific group.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `group` | string | No | Filter by Cognito group: `"admin"`, `"admins"`, `"user"`, or `"users"`. Omit to list all users. |
| `limit` | integer | No | Max users to return per page (Cognito limit) |
| `paginationToken` | string | No | Cognito pagination token from the previous response |
| `filter` | string | No | Cognito filter expression (e.g., `email = "user@example.com"`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "users": [
    {
      "id": "cognito-sub",
      "email": "jane@example.com",
      "name": "Jane Doe",
      "preferredUsername": "jane_v",
      "picture": "https://cdn.example.com/users/jane.jpg",
      "birthdate": "1990-05-20T00:00:00Z",
      "locale": "en",
      "userStatus": "CONFIRMED",
      "userType": "user",
      "enabled": true,
      "createdAt": "2024-01-01T08:00:00Z",
      "updatedAt": "2024-06-01T12:00:00Z"
    }
  ],
  "paginationToken": "cognito-next-page-token"
}
```

> **Note**: `paginationToken` is a Cognito-native cursor, not the same as the DynamoDB `nextToken` used by other list endpoints. Pass it as-is in subsequent requests.

**Error Responses**

| Status | Description |
|--------|-------------|
| `403 Forbidden` | Not a global admin |
| `500 Internal Server Error` | Cognito error |

---

### Get User

- **Endpoint**: `GET /api/v1/users/{userSub}`
- **Auth**: Global Admin only
- **Description**: Returns a user's full profile and all their team memberships.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `userSub` | string | Yes | Cognito sub (user ID) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "user": {
    "id": "cognito-sub",
    "email": "jane@example.com",
    "name": "Jane Doe",
    "preferredUsername": "jane_v",
    "picture": "https://cdn.example.com/users/jane.jpg",
    "birthdate": "1990-05-20T00:00:00Z",
    "locale": "en",
    "userStatus": "CONFIRMED",
    "userType": "user",
    "enabled": true,
    "createdAt": "2024-01-01T08:00:00Z",
    "updatedAt": "2024-06-01T12:00:00Z"
  },
  "memberships": [
    {
      "id": "member-uuid",
      "userId": "cognito-sub",
      "teamId": "team-uuid",
      "role": "trainer",
      "status": "active",
      "createdAt": "2024-02-01T09:00:00Z",
      "updatedAt": "2024-02-01T09:00:00Z",
      "joinedAt": "2024-02-01T09:00:00Z",
      "leftAt": null
    }
  ]
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `userSub` |
| `403 Forbidden` | Not a global admin |
| `404 Not Found` | User not found in Cognito |

---

### Update User

- **Endpoint**: `PATCH /api/v1/users/{userSub}`
- **Auth**: Global Admin only
- **Description**: Updates a user's type or enabled/disabled status. Promoting a user to `admin` automatically removes all their team memberships.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `userSub` | string | Yes | Cognito sub (user ID) |

**Request Body** (all fields optional)
```json
{
  "userType": "admin",
  "enabled": false
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `userType` | string | No | `"admin"` or `"user"`. Changing to `"admin"` removes all team memberships. |
| `enabled` | boolean | No | `true` to enable, `false` to disable the Cognito account |

**Response** `200 OK`
```json
{
  "message": "Success",
  "user": {
    "id": "cognito-sub",
    "email": "jane@example.com",
    "name": "Jane Doe",
    "userType": "admin",
    "enabled": false,
    "userStatus": "CONFIRMED"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `userSub` |
| `403 Forbidden` | Not a global admin |
| `404 Not Found` | User not found |

---

### Delete User

- **Endpoint**: `DELETE /api/v1/users/{userSub}`
- **Auth**: Global Admin only
- **Description**: Deletes a user from Cognito and removes all their team memberships.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `userSub` | string | Yes | Cognito sub (user ID) |

**Response** `200 OK`
```json
{
  "message": "Success"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `userSub` |
| `403 Forbidden` | Not a global admin |

---

## User Status Values (Cognito)

| Status | Description |
|--------|-------------|
| `CONFIRMED` | User confirmed their account |
| `FORCE_CHANGE_PASSWORD` | User must change their temporary password |
| `UNCONFIRMED` | Email not yet verified |
| `RESET_REQUIRED` | Admin-initiated password reset pending |

## User Type Values

| Type | Description |
|------|-------------|
| `user` | Regular team member (can belong to teams) |
| `admin` | Global administrator (cannot belong to teams) |
