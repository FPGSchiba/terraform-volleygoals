# Self API

The Self API allows the currently authenticated user to read and update their own profile. All endpoints are under `/api/v1/self`.

---

### Get Self

- **Endpoint**: `GET /api/v1/self`
- **Auth**: Any authenticated user
- **Description**: Returns the current user's profile and all their team assignments (team + role + status).

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
  "assignments": [
    {
      "team": {
        "id": "team-uuid",
        "name": "Volleyball Team A",
        "status": "active",
        "picture": "https://cdn.example.com/teams/team-uuid/logo.jpg",
        "createdAt": "2024-01-15T10:00:00Z",
        "updatedAt": "2024-06-01T12:00:00Z"
      },
      "role": "trainer",
      "status": "active"
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `user` | object | Full Cognito profile of the authenticated user |
| `assignments` | array | List of teams the user belongs to with their role and status |

**Error Responses**

| Status | Description |
|--------|-------------|
| `401 Unauthorized` | No authenticated user found |
| `404 Not Found` | User profile not found in Cognito |

---

### Update Self

- **Endpoint**: `PATCH /api/v1/self`
- **Auth**: Any authenticated user
- **Description**: Updates the current user's Cognito profile attributes. All fields are optional; only provided fields are updated.

**Request Body**
```json
{
  "name": "Jane Smith",
  "preferredUsername": "jane_smith",
  "birthdate": "1990-05-20",
  "language": "en"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | Display name |
| `preferredUsername` | string | No | Username / nickname |
| `birthdate` | string | No | Date of birth in `YYYY-MM-DD` format |
| `language` | string | No | Preferred language: `"en"` or `"de"` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "user": {
    "id": "cognito-sub",
    "email": "jane@example.com",
    "name": "Jane Smith",
    "preferredUsername": "jane_smith",
    "picture": "https://cdn.example.com/users/jane.jpg",
    "birthdate": "1990-05-20T00:00:00Z",
    "locale": "en",
    "userStatus": "CONFIRMED",
    "userType": "user",
    "enabled": true,
    "createdAt": "2024-01-01T08:00:00Z",
    "updatedAt": "2024-11-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body |
| `401 Unauthorized` | Not authenticated |
| `404 Not Found` | User not found after update |

---

### Upload Profile Picture (Presign)

- **Endpoint**: `GET /api/v1/self/picture/presign`
- **Auth**: Any authenticated user
- **Description**: Generates a presigned S3 URL to upload a profile picture. After uploading directly to S3, the user's Cognito `picture` attribute is automatically updated to the public URL.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filename` | string | Yes | Original filename (e.g., `avatar.jpg`) |
| `contentType` | string | Yes | MIME type (e.g., `image/jpeg`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "uploadUrl": "https://s3.amazonaws.com/bucket/users/cognito-sub/avatar.jpg?X-Amz-...",
  "key": "users/cognito-sub/avatar.jpg",
  "fileUrl": "https://cdn.example.com/users/cognito-sub/avatar.jpg"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `uploadUrl` | string | Pre-signed S3 PUT URL — upload directly using HTTP PUT |
| `key` | string | Storage key in S3 |
| `fileUrl` | string | Public URL that will be set as the user's `picture` attribute |

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `filename` or `contentType` |
| `401 Unauthorized` | Not authenticated |
