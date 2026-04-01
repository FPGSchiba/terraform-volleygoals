# Tenants API

Tenants represent organizations that own and manage multiple teams. The Tenant API covers tenant management, role definitions, ownership policies, and creating teams within a tenant.

All endpoints are under `/v1/tenants`.

---

## Tenants

### Create Tenant

- **Endpoint**: `POST /v1/tenants`
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
  "data": {
    "tenant": {
      "id": "tenant-uuid",
      "name": "Sports Club Zurich",
      "ownerId": "cognito-sub",
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z"
    }
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing or empty `name` |
| `403 Forbidden` | Not a global admin |

---

### Get Tenant

- **Endpoint**: `GET /v1/tenants/{tenantId}`
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
  "data": {
    "tenant": {
      "id": "tenant-uuid",
      "name": "Sports Club Zurich",
      "ownerId": "cognito-sub",
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-01-01T10:00:00Z"
    }
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

- **Endpoint**: `PATCH /v1/tenants/{tenantId}`
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
  "data": {
    "tenant": {
      "id": "tenant-uuid",
      "name": "Sports Club Geneva",
      "ownerId": "cognito-sub",
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-11-01T10:00:00Z"
    }
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

- **Endpoint**: `DELETE /v1/tenants/{tenantId}`
- **Auth**: Global Admin only
- **Description**: Deletes a tenant and all associated data.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK`
```json
{
  "message": "Success",
  "data": null
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

### Add Tenant Member

- **Endpoint**: `POST /v1/tenants/{tenantId}/members`
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
  "data": {
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
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing or invalid body |
| `403 Forbidden` | Not authorized |

---

### Remove Tenant Member

- **Endpoint**: `DELETE /v1/tenants/{tenantId}/members/{memberId}`
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
  "message": "Success",
  "data": null
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

- **Endpoint**: `GET /v1/tenants/{tenantId}/roles`
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
  "data": {
    "roles": [
      {
        "id": "role-uuid",
        "tenantId": "tenant-uuid",
        "name": "Coach",
        "permissions": ["goals:read", "goals:write", "progress_reports:read"],
        "isDefault": false,
        "createdAt": "2024-01-01T10:00:00Z",
        "updatedAt": "2024-01-01T10:00:00Z"
      }
    ]
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Missing `tenantId` |
| `403 Forbidden` | Not authorized |

---

### Create Role Definition

- **Endpoint**: `POST /v1/tenants/{tenantId}/roles`
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
  "data": {
    "role": {
      "id": "role-uuid",
      "tenantId": "tenant-uuid",
      "name": "Coach",
      "permissions": ["goals:read", "goals:write", "progress_reports:read", "comments:read"],
      "isDefault": false,
      "createdAt": "2024-06-01T10:00:00Z",
      "updatedAt": "2024-06-01T10:00:00Z"
    }
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

- **Endpoint**: `PATCH /v1/tenants/{tenantId}/roles/{roleId}`
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
  "data": {
    "role": {
      "id": "role-uuid",
      "tenantId": "tenant-uuid",
      "name": "Coach",
      "permissions": ["goals:read", "goals:write", "progress_reports:read", "progress_reports:write"],
      "isDefault": false,
      "createdAt": "2024-06-01T10:00:00Z",
      "updatedAt": "2024-11-01T10:00:00Z"
    }
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

- **Endpoint**: `DELETE /v1/tenants/{tenantId}/roles/{roleId}`
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
  "message": "Success",
  "data": null
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

- **Endpoint**: `GET /v1/tenants/{tenantId}/ownership-policies`
- **Auth**: Global Admin or tenant admin
- **Description**: Returns all ownership policies for a tenant.

**Path Parameters**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tenantId` | string (UUID) | Yes | ID of the tenant |

**Response** `200 OK`
```json
{
  "message": "Success",
  "data": {
    "policies": [
      {
        "id": "policy-uuid",
        "tenantId": "tenant-uuid",
        "resourceType": "goals",
        "ownerPermissions": ["goals:read", "goals:write", "goals:delete"],
        "parentOwnerPermissions": ["goals:read"],
        "createdAt": "2024-01-01T10:00:00Z",
        "updatedAt": "2024-01-01T10:00:00Z"
      }
    ]
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

- **Endpoint**: `PATCH /v1/tenants/{tenantId}/ownership-policies/{resourceType}`
- **Auth**: Global Admin or tenant admin
- **Description**: Creates or updates the ownership policy for a specific resource type within a tenant (upsert behavior).

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
  "data": {
    "policy": {
      "id": "policy-uuid",
      "tenantId": "tenant-uuid",
      "resourceType": "goals",
      "ownerPermissions": ["goals:read", "goals:write", "goals:delete"],
      "parentOwnerPermissions": ["goals:read", "goals:write"],
      "createdAt": "2024-01-01T10:00:00Z",
      "updatedAt": "2024-11-01T10:00:00Z"
    }
  }
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing path params |
| `403 Forbidden` | Not authorized |

---

## Tenanted Teams

### Create Tenanted Team

- **Endpoint**: `POST /v1/tenants/{tenantId}/teams`
- **Auth**: Global Admin or tenant admin
- **Description**: Creates a new team that is automatically linked to the given tenant. This is the preferred way to create teams within a tenant instead of the standalone `POST /v1/teams`.

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
  "data": {
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
}
```

**Error Responses**

| Status | Description |
|--------|-------------|
| `400 Bad Request` | Invalid body or missing `tenantId` |
| `403 Forbidden` | Not authorized |
| `404 Not Found` | Tenant not found |
