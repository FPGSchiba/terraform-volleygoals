# Tenants API

Tenants represent organizations that own and manage multiple teams. The Tenant API covers tenant management, role definitions, ownership policies, and creating teams within a tenant.

All endpoints are under `/api/v1/tenants`.

---

## Resource Definitions

### Get Resource Definitions

- **Endpoint**: `GET /api/v1/resource-definitions`
- **Auth**: Authenticated caller; intended for UI rendering and tenant-admin tooling
- **Description**: Returns the static resource definitions used by the frontend to render resources, actions, and allowed child resources dynamically.

**Response** `200 OK`
```json
{
  "message": "Success",
  "data": [
    {
      "id": "goals",
      "name": "Goals",
      "description": "Individual or team goals",
      "actions": ["read", "write", "delete"],
      "allowedChildResources": ["comments", "progressReports", "seasons"]
    },
    {
      "id": "comments",
      "name": "Comments",
      "description": "Comments attached to goals or progress reports",
      "actions": ["read", "write", "delete"],
      "allowedChildResources": []
    }
  ]
}
```

**Notes**

- Each resource definition includes a non-empty `actions` array.
- The current implementation returns the array payload under `data`.

---

## Resource Model

### Get Resource Model

- **Endpoint**: `GET /api/v1/tenants/{tenantId}/resource-model`
- **Auth**: Global Admin or tenant admin
- **Description**: Returns resource definitions and effective ownership policies in one request so the UI can fetch everything it needs for permission editing.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK`
```json
{
  "message": "Success",
  "resourceDefinitions": [
    {
      "id": "goals",
      "name": "Goals",
      "description": "Individual or team goals",
      "actions": ["read", "write", "delete"],
      "allowedChildResources": ["comments", "progressReports", "seasons"]
    }
  ],
  "policies": [
    {
      "id": "policy-uuid",
      "tenantId": "tenant-uuid",
      "resourceType": "goals",
      "ownerPermissions": ["goals:read", "comments:read"],
      "parentOwnerPermissions": ["goals:read"]
    }
  ]
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized |

---

## Tenants

### List Tenants

- **Endpoint**: `GET /api/v1/tenants`
- **Auth**: Global Admin only
- **Description**: Returns a paginated list of all tenants.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |
| `name` | string | No | Filter by tenant name (partial match) |
| `ownerId` | string | No | Filter by tenant owner Cognito sub |
| `createdAfter` | string (RFC3339) | No | Return tenants created on or after this timestamp |
| `createdBefore` | string (RFC3339) | No | Return tenants created on or before this timestamp |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "tenant-uuid",
      "name": "Sports Club Zurich",
      "ownerId": "cognito-sub",
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z"
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
| `403 Forbidden` | Not a global admin |

---

### Create Tenant

- **Endpoint**: `POST /api/v1/tenants`
- **Auth**: Global Admin only
- **Description**: Creates a new tenant. The calling admin becomes the initial owner.

**Request Body**
```json
{
  "name": "Sports Club Zurich"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Tenant/organization name (must not be empty) |

**Response** `201 Created`
```json
{
  "message": "Success",
  "tenant": {
    "id": "tenant-uuid",
    "name": "Sports Club Zurich",
    "ownerId": "cognito-sub",
    "createdAt": "2024-01-01T10:00:00Z",
    "updatedAt": "2024-01-01T10:00:00Z"
  }
}
```

**Notes**

- After the tenant is created, the calling admin is automatically added as a `tenant_admin` member of the new tenant. This ensures the creator always has administrative access without a separate `POST /api/v1/tenants/{tenantId}/members` call.

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing or empty `name` |
| `403 Forbidden` | Not a global admin |

---

### Get Tenant

- **Endpoint**: `GET /api/v1/tenants/{tenantId}`
- **Auth**: Global Admin, or an active member of the tenant
- **Description**: Returns a tenant by ID.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK`
```json
{
  "message": "Success",
  "tenant": {
    "id": "tenant-uuid",
    "name": "Sports Club Zurich",
    "ownerId": "cognito-sub",
    "createdAt": "2024-01-01T10:00:00Z",
    "updatedAt": "2024-01-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized (not admin, not an active tenant member) |
| `404 Not Found` | Tenant not found |

---

### Update Tenant

- **Endpoint**: `PATCH /api/v1/tenants/{tenantId}`
- **Auth**: Global Admin or tenant admin
- **Description**: Updates the tenant's name.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Request Body**
```json
{
  "name": "Sports Club Geneva"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | No | New tenant name |

**Response** `200 OK`
```json
{
  "message": "Success",
  "tenant": {
    "id": "tenant-uuid",
    "name": "Sports Club Geneva",
    "ownerId": "cognito-sub",
    "createdAt": "2024-01-01T10:00:00Z",
    "updatedAt": "2024-11-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `tenantId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Tenant not found |

---

### Delete Tenant

- **Endpoint**: `DELETE /api/v1/tenants/{tenantId}`
- **Auth**: Global Admin only
- **Description**: Deletes a tenant and all associated data.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK`
```json
{
  "message": "Success"
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not a global admin |
| `404 Not Found` | Tenant not found |

---

## Tenant Members

### List Tenant Members

- **Endpoint**: `GET /api/v1/tenants/{tenantId}/members`
- **Auth**: Global Admin or tenant admin
- **Description**: Returns a paginated list of members for a specific tenant.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |
| `role` | string | No | Filter by tenant member role (`tenant_admin`, `tenant_member`) |
| `userId` | string | No | Filter by user Cognito sub |
| `status` | string | No | Filter by member status (`active`, `removed`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "tenant-uuid#cognito-sub",
      "tenantId": "tenant-uuid",
      "userId": "cognito-sub",
      "role": "tenant_admin",
      "status": "active",
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z"
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
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized |

---

### Add Tenant Member

- **Endpoint**: `POST /api/v1/tenants/{tenantId}/members`
- **Auth**: Global Admin or tenant admin
- **Description**: Adds an existing user to a tenant with a specified role.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

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
| `role` | string | No | `"admin"` or `"member"` (defaults to `"member"`) |

**Response** `201 Created`
```json
{
  "message": "Success",
  "member": {
    "id": "member-uuid",
    "tenantId": "tenant-uuid",
    "userId": "cognito-sub",
    "role": "member",
    "status": "active",
    "createdAt": "2024-06-01T10:00:00Z",
    "updatedAt": "2024-06-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing or invalid body |
| `403 Forbidden` | Not authorized |

---

### Remove Tenant Member

- **Endpoint**: `DELETE /api/v1/tenants/{tenantId}/members/{memberId}`
- **Auth**: Global Admin or tenant admin
- **Description**: Removes a member from a tenant.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |
| `memberId` | string (UUID) | Yes | ID of the tenant member record |

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
| `404 Not Found` | Member not found or not in this tenant |

---

## Role Definitions

Role definitions define custom permission sets for team members within a tenant.

### List Role Definitions

- **Endpoint**: `GET /api/v1/tenants/{tenantId}/roles`
- **Auth**: Global Admin or tenant admin
- **Description**: Returns all role definitions for a tenant.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK`
```json
{
  "message": "Success",
  "roles": [
    {
      "id": "role-uuid",
      "tenantId": "tenant-uuid",
      "name": "Coach",
      "permissions": [
        "goals:read",
        "goals:write",
        "progress_reports:read"
      ],
      "isDefault": false,
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z"
    }
  ]
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized |

---

### Create Role Definition

- **Endpoint**: `POST /api/v1/tenants/{tenantId}/roles`
- **Auth**: Global Admin or tenant admin
- **Description**: Creates a new custom role definition for a tenant.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Request Body**
```json
{
  "name": "Coach",
  "permissions": ["goals:read", "goals:write", "progress_reports:read", "comments:read"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Role display name (must not be empty) |
| `permissions` | array of strings | Yes | List of permission strings (must not be empty) |

**Response** `201 Created`
```json
{
  "message": "Success",
  "role": {
    "id": "role-uuid",
    "tenantId": "tenant-uuid",
    "name": "Coach",
    "permissions": [
      "goals:read",
      "goals:write",
      "progress_reports:read",
      "comments:read"
    ],
    "isDefault": false,
    "createdAt": "2024-06-01T10:00:00Z",
    "updatedAt": "2024-06-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body, empty name, or empty permissions |
| `403 Forbidden` | Not authorized |

---

### Update Role Definition

- **Endpoint**: `PATCH /api/v1/tenants/{tenantId}/roles/{roleId}`
- **Auth**: Global Admin or tenant admin
- **Description**: Updates the permissions of a custom role. Default roles (`isDefault: true`) cannot be modified.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |
| `roleId` | string (UUID) | Yes | ID of the role definition |

**Request Body**
```json
{
  "permissions": ["goals:read", "goals:write", "progress_reports:read", "progress_reports:write"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `permissions` | array of strings | Yes | New complete list of permissions (replaces existing) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "role": {
    "id": "role-uuid",
    "tenantId": "tenant-uuid",
    "name": "Coach",
    "permissions": [
      "goals:read",
      "goals:write",
      "progress_reports:read",
      "progress_reports:write"
    ],
    "isDefault": false,
    "createdAt": "2024-06-01T10:00:00Z",
    "updatedAt": "2024-11-01T10:00:00Z"
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing path params |
| `403 Forbidden` | Not authorized or role is a default role |
| `404 Not Found` | Role not found in this tenant |

---

### Delete Role Definition

- **Endpoint**: `DELETE /api/v1/tenants/{tenantId}/roles/{roleId}`
- **Auth**: Global Admin or tenant admin
- **Description**: Deletes a custom role definition. Default roles cannot be deleted.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |
| `roleId` | string (UUID) | Yes | ID of the role definition |

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
| `403 Forbidden` | Not authorized or role is a default role |
| `404 Not Found` | Role not found in this tenant |

---

## Ownership Policies

Ownership policies define what permissions are automatically granted to the **owner** of a resource and to the **parent resource owner** (e.g., the owner of the season that contains a goal).

### List Ownership Policies

- **Endpoint**: `GET /api/v1/tenants/{tenantId}/ownership-policies`
- **Auth**: Global Admin or tenant admin
- **Description**: Returns the normalized ownership policy list for a tenant. The response always includes one entry per resource type (`goals`, `comments`, `progressReports`, `progress`, `seasons`). Missing policies are represented with empty/null permission fields.

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `format` | string | No | Use `nested` to return the legacy nested object shape for backward compatibility |

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK` (default normalized list)
```json
{
  "message": "Success",
  "policies": [
    {
      "id": "policy-uuid",
      "tenantId": "tenant-uuid",
      "resourceType": "goals",
      "ownerPermissions": ["goals:read", "comments:read"],
      "parentOwnerPermissions": null
    }
  ]
}
```

**Legacy Response** `200 OK` with `?format=nested`
```json
{
  "message": "Success",
  "policies": {
    "goals": {
      "resourceType": "goals",
      "ownerPermissions": ["goals:read", "comments:read"],
      "parentOwnerPermissions": null
    }
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized |

---

### Update Ownership Policy

- **Endpoint**: `PATCH /api/v1/tenants/{tenantId}/ownership-policies/{resourceType}`
- **Auth**: Global Admin or tenant admin
- **Description**: Creates or updates the ownership policy for a specific resource type within a tenant (upsert behavior).
- **Validation**: Each permission must use the `<resource>:<action>` format, and the action must exist in `resource-definitions`.
- **Global override**: If `tenantId` is `global`, provide the target tenant via the `X-Target-Tenant` header or the `targetTenant` query parameter.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |
| `resourceType` | string | Yes | Resource type (e.g., `goals`, `progress_reports`, `comments`) |

**Request Body**
```json
{
  "ownerPermissions": ["goals:read", "goals:write", "goals:delete"],
  "parentOwnerPermissions": ["goals:read", "goals:write"]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ownerPermissions` | array of strings | Yes | Permissions auto-granted to the resource owner |
| `parentOwnerPermissions` | array of strings | Yes | Permissions auto-granted to the owner of the parent resource |

**Response** `200 OK`
```json
{
  "message": "Success",
  "policy": {
    "id": "policy-uuid",
    "tenantId": "tenant-uuid",
    "resourceType": "goals",
    "ownerPermissions": [
      "goals:read",
      "goals:write",
      "goals:delete"
    ],
    "parentOwnerPermissions": [
      "goals:read",
      "goals:write"
    ],
    "createdAt": "2024-01-01T10:00:00Z",
    "updatedAt": "2024-11-01T10:00:00Z"
  }
}
```

**Notes**

- The saved policy is returned after validation and upsert.

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing path params |
| `403 Forbidden` | Not authorized |

---

### Batch Upsert Ownership Policies

- **Endpoint**: `PUT /api/v1/tenants/{tenantId}/ownership-policies`
- **Auth**: Global Admin or tenant admin
- **Description**: Upserts multiple ownership policies in one request. The implementation uses a best-effort compensation strategy if one of the upserts fails.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Request Body**
```json
[
  {
    "resourceType": "goals",
    "ownerPermissions": ["goals:read", "comments:read"],
    "parentOwnerPermissions": ["goals:read"]
  },
  {
    "resourceType": "comments",
    "ownerPermissions": ["comments:read", "comments:write"],
    "parentOwnerPermissions": []
  }
]
```

**Response** `200 OK`
```json
{
  "message": "Success",
  "policies": [
    {
      "id": "policy-uuid",
      "tenantId": "tenant-uuid",
      "resourceType": "goals",
      "ownerPermissions": ["goals:read", "comments:read"],
      "parentOwnerPermissions": ["goals:read"]
    }
  ]
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body, missing `tenantId`, or invalid permissions |
| `403 Forbidden` | Not authorized |

---

### Preview Effective Permissions

- **Endpoint**: `POST /api/v1/tenants/{tenantId}/ownership-policies/preview`
- **Auth**: Global Admin or tenant admin
- **Description**: Computes the effective permissions granted by the ownership policy for a resource type.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Request Body**
```json
{
  "resourceType": "goals",
  "resourceId": "goal-uuid"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `resourceType` | string | Yes | Resource type to preview |
| `resourceId` | string | No | Optional resource ID for future expansion |

**Response** `200 OK`
```json
{
  "message": "Success",
  "effectivePermissions": ["goals:read", "comments:read"],
  "sources": [
    {
      "permission": "goals:read",
      "sourceResourceType": "goals",
      "policyId": "policy-uuid"
    }
  ]
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body, missing `tenantId`, or unknown `resourceType` |
| `403 Forbidden` | Not authorized |

---

## Tenanted Teams

### List Tenanted Teams

- **Endpoint**: `GET /api/v1/tenants/{tenantId}/teams`
- **Auth**: Global Admin or tenant admin
- **Description**: Returns a paginated list of teams assigned to a tenant.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Query Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `limit` | integer | No | Items per page |
| `nextToken` | string | No | Pagination cursor |
| `sortBy` | string | No | Field to sort by |
| `sortOrder` | string | No | `asc` or `desc` |
| `name` | string | No | Filter by team name (partial match) |
| `status` | string | No | Filter by team status (`active`, `inactive`) |

**Response** `200 OK`
```json
{
  "message": "Success",
  "items": [
    {
      "id": "team-uuid",
      "name": "U21 Team",
      "status": "active",
      "picture": "",
      "tenantId": "tenant-uuid",
      "createdAt": "2024-06-01T10:00:00Z",
      "updatedAt": "2024-06-01T10:00:00Z",
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
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Tenant not found |

---

### Create Tenanted Team

- **Endpoint**: `POST /api/v1/tenants/{tenantId}/teams`
- **Auth**: Global Admin or tenant admin
- **Description**: Creates a new team that is automatically linked to the given tenant. This is the preferred way to create teams within a tenant instead of the standalone `POST /api/v1/teams`.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Request Body**
```json
{
  "name": "U21 Team"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Team name (must not be empty) |

**Response** `201 Created`
```json
{
  "message": "Success",
  "team": {
    "id": "team-uuid",
    "name": "U21 Team",
    "status": "active",
    "picture": "",
    "tenantId": "tenant-uuid",
    "createdAt": "2024-06-01T10:00:00Z",
    "updatedAt": "2024-06-01T10:00:00Z",
    "deletedAt": null
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `tenantId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Tenant not found |
