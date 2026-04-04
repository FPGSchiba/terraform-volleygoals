# Comments API

Comments can be attached to Goals, Progress Reports, or Progress Entries. All comment endpoints are under `/api/v1/comments`.

---

### Create Comment

- **Endpoint**: `POST /api/v1/comments`
- **Auth**: Team member with `comments:write` permission (checked against the parent resource)
- **Description**: Creates a comment on a goal, progress report, or progress entry. Whether goal comments are allowed depends on the team's settings (`allowTeamGoalComments`, `allowIndividualGoalComments`).

**Request Body**
```json
{
  "commentType": "Goal",
  "targetId": "goal-uuid",
  "content": "Great work on this goal, keep it up!"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `commentType` | string | Yes | `"Goal"`, `"ProgressReport"`, or `"ProgressEntry"` |
| `targetId` | string (UUID) | Yes | ID of the resource being commented on |
| `content` | string | Yes | Comment text (must not be empty) |

**Response** `201 Created`
```json
{
  "message": "Success",
  "comment": {
    "id": "comment-uuid",
    "authorId": "cognito-sub",
    "authorName": "Jane Doe",
    "authorPicture": "https://cdn.example.com/users/jane.jpg",
    "commentType": "Goal",
    "targetId": "goal-uuid",
    "content": "Great work on this goal, keep it up!",
    "createdAt": "2024-10-05T10:00:00Z",
    "updatedAt": "2024-10-05T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing or invalid fields |
| `403 Forbidden` | Not authorized, or comments disabled in team settings |
| `404 Not Found` | Target resource not found |

---

### Get Comment

- **Endpoint**: `GET /api/v1/comments/{commentId}`
- **Auth**: Team member with `comments:read` permission (ownership-aware)
- **Description**: Returns a single comment and all its attached files.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `commentId` | string (UUID) | Yes | ID of the comment |

**Response** `200 OK`
```json
{
  "message": "Success",
  "comment": {
    "id": "comment-uuid",
    "authorId": "cognito-sub",
    "authorName": "Jane Doe",
    "authorPicture": "https://cdn.example.com/users/jane.jpg",
    "commentType": "Goal",
    "targetId": "goal-uuid",
    "content": "Great work on this goal, keep it up!",
    "createdAt": "2024-10-05T10:00:00Z",
    "updatedAt": "2024-10-05T10:00:00Z"
  },
  "files": [
    {
      "id": "file-uuid",
      "commentId": "comment-uuid",
      "storageKey": "comments/comment-uuid/attachment.pdf",
      "fileUrl": "https://cdn.example.com/comments/comment-uuid/attachment.pdf"
    }
  ]
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `commentId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Comment not found |

---

### List Comments

- **Endpoint**: `GET /api/v1/comments`
- **Auth**: Team member with `comments:read` permission
- **Description**: Returns a paginated list of comments for a specific target resource.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `commentType` | string | **Yes** | `"Goal"`, `"ProgressReport"`, or `"ProgressEntry"` |
| `targetId` | string (UUID) | **Yes** | ID of the resource whose comments to list |
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
      "id": "comment-uuid",
      "authorId": "cognito-sub",
      "authorName": "Jane Doe",
      "authorPicture": "https://cdn.example.com/users/jane.jpg",
      "commentType": "Goal",
      "targetId": "goal-uuid",
      "content": "Great work!",
      "createdAt": "2024-10-05T10:00:00Z",
      "updatedAt": "2024-10-05T10:00:00Z"
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
| `400 Bad Request` | Missing `commentType` or `targetId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Target resource not found |

---

### Update Comment

- **Endpoint**: `PATCH /api/v1/comments/{commentId}`
- **Auth**: Comment author or team member with `comments:write` permission for the author's comments
- **Description**: Updates the text content of a comment.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `commentId` | string (UUID) | Yes | ID of the comment |

**Request Body**
```json
{
  "content": "Updated comment text"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | string | Yes | New comment text |

**Response** `200 OK`
```json
{
  "message": "Success",
  "comment": {
    "id": "comment-uuid",
    "authorId": "cognito-sub",
    "authorName": "Jane Doe",
    "commentType": "Goal",
    "targetId": "goal-uuid",
    "content": "Updated comment text",
    "createdAt": "2024-10-05T10:00:00Z",
    "updatedAt": "2024-10-06T08:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `commentId` |
| `403 Forbidden` | Not authorized (not the author or lacking permission) |
| `404 Not Found` | Comment not found |

---

### Delete Comment

- **Endpoint**: `DELETE /api/v1/comments/{commentId}`
- **Auth**: Comment author or team member with `comments:delete` permission for the author's comments
- **Description**: Deletes a comment and all its attached files.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `commentId` | string (UUID) | Yes | ID of the comment |

**Response** `204 No Content`

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `commentId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Comment not found |

---

### Upload Comment File (Presign)

- **Endpoint**: `GET /api/v1/comments/{commentId}/file/presign`
- **Auth**: Comment author or team member with `comments:write` for the author's comments
- **Description**: Generates a presigned S3 upload URL to attach a file to a comment. A `CommentFile` record is created immediately with the storage key.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `commentId` | string (UUID) | Yes | ID of the comment |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `filename` | string | Yes | Original filename (e.g., `document.pdf`) |
| `contentType` | string | Yes | MIME type (e.g., `application/pdf`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "uploadUrl": "https://s3.amazonaws.com/bucket/comments/comment-uuid/document.pdf?X-Amz-...",
  "key": "comments/comment-uuid/document.pdf",
  "commentFile": {
    "id": "file-uuid",
    "commentId": "comment-uuid",
    "storageKey": "comments/comment-uuid/document.pdf"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing path param, `filename`, or `contentType` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Comment not found |

---

## Comment Types

| Type | Target |
|------|--------|
| `Goal` | A goal record |
| `ProgressReport` | A progress report record |
| `ProgressEntry` | An individual progress entry within a report |
