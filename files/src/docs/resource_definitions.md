# Canonical Resource Definitions

This document describes the canonical resource definitions used by the backend as the single source of truth for permission strings, allowed actions, and child relations.

Location
- The programmatic definitions are provided by `models.GetDefinitions()` in `files/src/models/resource_definitions.go`.

Response shape (GET /resource-definitions)

The endpoint (if implemented) or the export script below returns a JSON array of objects with the following fields:

- `id` (string): underscore-style resource ID used for permissions, e.g. `progress_reports`.
- `name` (string): human-friendly name.
- `description` (string): short description of the resource.
- `actions` ([]string): allowed actions for the resource (e.g. `read`, `write`, `delete`).
- `allowedChildResources` ([]string): IDs of allowed child resource types.

Example JSON (abridged)

```json
[
  {
    "id": "goals",
    "name": "Goals",
    "description": "Individual or team goals",
    "actions": ["read", "write", "delete"],
    "allowedChildResources": ["comments", "progress_reports", "goal_seasons", "activities"]
  },
  {
    "id": "comments",
    "name": "Comments",
    "description": "Comments attached to goals or progress reports",
    "actions": ["read", "write", "delete"],
    "allowedChildResources": ["comment_files", "activities"]
  }
]
```

Notes for frontend
- Use the `id` field to generate permission strings with the action (e.g. `${id}:read`).
- Actions are authoritative per resource; validate client-side forms against the `actions` list to prevent invalid requests.
- If you need a JSON file for build-time consumption, run the small export script at `files/src/scripts/export_resource_definitions/main.go`.

