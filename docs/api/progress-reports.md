# Progress Reports API

Progress Reports are nested under seasons: `/api/v1/seasons/{seasonId}/progress-reports`.

A progress report is a periodic self-assessment submitted by a team member. It includes an overall summary and individual progress ratings per goal (progress entries).

---

### Create Progress Report

- **Endpoint**: `POST /api/v1/seasons/{seasonId}/progress-reports`
- **Auth**: Team member with `progress_reports:write` permission
- **Description**: Creates a new progress report for the caller. Emits a `progress_report.created` activity event.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the report belongs to |

**Request Body**
```json
{
  "summary": "Good progress this week",
  "details": "Focused on serving drills every day",
  "overallDetails": "Overall I'm improving but need to work on defense",
  "progress": [
    {
      "goalId": "goal-uuid",
      "rating": 4,
      "details": "Achieved 75% accuracy in practice"
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `summary` | string | Yes | Short one-line summary |
| `details` | string | Yes | Detailed notes for the report |
| `overallDetails` | string | Yes | Overall reflection on progress |
| `progress` | array | No | Array of goal progress entries |
| `progress[].goalId` | string (UUID) | Yes (in entry) | Goal being rated |
| `progress[].rating` | integer (1–5) | Yes (in entry) | Progress rating from 1 (low) to 5 (high) |
| `progress[].details` | string | No | Notes for this specific goal |

**Response** `201 Created`
```json
{
  "message": "Success",
  "progressReport": {
    "id": "report-uuid",
    "seasonId": "season-uuid",
    "authorId": "cognito-sub",
    "authorName": "Jane Doe",
    "authorPicture": "https://cdn.example.com/users/jane.jpg",
    "summary": "Good progress this week",
    "details": "Focused on serving drills every day",
    "overallDetails": "Overall I'm improving but need to work on defense",
    "createdAt": "2024-10-01T09:00:00Z",
    "updatedAt": "2024-10-01T09:00:00Z",
    "progress": [
      {
        "id": "progress-uuid",
        "progressReportId": "report-uuid",
        "goalId": "goal-uuid",
        "rating": 4,
        "details": "Achieved 75% accuracy in practice"
      }
    ]
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `seasonId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |

---

### Get Progress Report

- **Endpoint**: `GET /api/v1/seasons/{seasonId}/progress-reports/{reportId}`
- **Auth**: Team member with `progress_reports:read` permission (ownership-aware)
- **Description**: Returns a single progress report with all its progress entries.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the report belongs to |
| `reportId` | string (UUID) | Yes | ID of the progress report |

**Response** `200 OK`
```json
{
  "message": "Success",
  "progressReport": {
    "id": "report-uuid",
    "seasonId": "season-uuid",
    "authorId": "cognito-sub",
    "authorName": "Jane Doe",
    "authorPicture": "https://cdn.example.com/users/jane.jpg",
    "summary": "Good progress this week",
    "details": "Focused on serving drills every day",
    "overallDetails": "Overall I'm improving but need to work on defense",
    "createdAt": "2024-10-01T09:00:00Z",
    "updatedAt": "2024-10-01T09:00:00Z",
    "progress": [
      {
        "id": "progress-uuid",
        "progressReportId": "report-uuid",
        "goalId": "goal-uuid",
        "rating": 4,
        "details": "Achieved 75% accuracy in practice"
      }
    ]
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing path params |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Report not found in the given season |

---

### List Progress Reports

- **Endpoint**: `GET /api/v1/seasons/{seasonId}/progress-reports`
- **Auth**: Team member with `progress_reports:read` permission
- **Description**: Returns a paginated list of progress reports for a season, each enriched with progress entries.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season to list reports for |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `authorId` | string | No | Filter by author's Cognito sub |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "report-uuid",
      "seasonId": "season-uuid",
      "authorId": "cognito-sub",
      "authorName": "Jane Doe",
      "authorPicture": "https://cdn.example.com/users/jane.jpg",
      "summary": "Good progress this week",
      "details": "Focused on serving drills every day",
      "overallDetails": "Overall I'm improving but need to work on defense",
      "createdAt": "2024-10-01T09:00:00Z",
      "updatedAt": "2024-10-01T09:00:00Z",
      "progress": [
        {
          "id": "progress-uuid",
          "progressReportId": "report-uuid",
          "goalId": "goal-uuid",
          "rating": 4,
          "details": "Achieved 75% accuracy in practice"
        }
      ]
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
| `400 Bad Request` | Missing `seasonId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Season not found |

---

### Update Progress Report

- **Endpoint**: `PATCH /api/v1/seasons/{seasonId}/progress-reports/{reportId}`
- **Auth**: Report author or team member with `progress_reports:write` on the author's reports
- **Description**: Updates a progress report's summary, details, or progress entries. Providing `progress` replaces all existing entries.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the report belongs to |
| `reportId` | string (UUID) | Yes | ID of the progress report |

**Request Body** (all fields optional)
```json
{
  "summary": "Updated summary",
  "details": "Updated details",
  "overallDetails": "Updated overall reflection",
  "progress": [
    {
      "goalId": "goal-uuid",
      "rating": 5,
      "details": "Exceeded expectations"
    }
  ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `summary` | string | No | Updated summary |
| `details` | string | No | Updated details |
| `overallDetails` | string | No | Updated overall reflection |
| `progress` | array | No | Replaces all progress entries |

**Response** `200 OK`
```json
{
  "message": "Success",
  "progressReport": {
    "id": "report-uuid",
    "seasonId": "season-uuid",
    "authorId": "cognito-sub",
    "authorName": "Jane Doe",
    "summary": "Updated summary",
    "details": "Updated details",
    "overallDetails": "Updated overall reflection",
    "createdAt": "2024-10-01T09:00:00Z",
    "updatedAt": "2024-10-05T11:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing path params |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Report not found in the given season |

---

### Delete Progress Report

- **Endpoint**: `DELETE /api/v1/seasons/{seasonId}/progress-reports/{reportId}`
- **Auth**: Report author or team member with `progress_reports:delete` on the author's reports
- **Description**: Deletes a progress report and all its progress entries.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `seasonId` | string (UUID) | Yes | Season the report belongs to |
| `reportId` | string (UUID) | Yes | ID of the progress report |

**Response** `204 No Content`

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing path params |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Report not found in the given season |

---

## Progress Entry Ratings

| Rating | Meaning |
|--------|---------|
| `1` | No progress |
| `2` | Minimal progress |
| `3` | Moderate progress |
| `4` | Good progress |
| `5` | Goal fully achieved |

The `completionPercentage` on goals is calculated as: `round((avg_rating / 5) * 100)`.
