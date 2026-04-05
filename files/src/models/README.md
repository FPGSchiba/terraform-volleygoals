# Resource Definitions — canonical model

This file documents the canonical resource definitions used by the backend. The authoritative data is returned by `models.GetDefinitions()` (see `resource_definitions.go`).

Purpose
- Provide a single source of truth for resource IDs, available actions, and allowed child resources.
- Ensure permission strings are generated consistently (resource:action) and use underscore-style IDs.

ResourceDefinition format
- Each resource is represented by the `ResourceDefinition` struct (see `resource_definition.go`):
  - `Id` (string): underscore-style identifier used in permission strings, e.g. `progress_reports`.
  - `Name` (string): human-friendly name for UI.
  - `Description` (string): short description of the resource.
  - `Actions` ([]string): list of supported actions (commonly `read`, `write`, `delete`).
  - `AllowedChildResources` ([]string): child resource IDs that may be created/owned under this resource.

ID naming conventions
- IDs MUST use lowercase underscore_style for multi-word names (e.g. `team_settings`, not `teamSettings`).
- IDs are the canonical values used in permission strings: `<resource_id>:<action>` (for example `comments:read`).

Usage examples
- Get the canonical definitions:

  defs := models.GetDefinitions()

- Build a permission string for a resource and action (example):

  // helper
  perm := func(resource, action string) string { return fmt.Sprintf("%s:%s", resource, action) }
  p := perm("goals", "read") // "goals:read"

- Prefer using generated helpers (when available) rather than constructing strings manually.

Migration / TODOs
- If you find router files that define local copies of resource definitions or hard-coded permission strings in non-underscore style, migrate them to call `models.GetDefinitions()` and derive permissions from it.
- Recommended marker for files that still contain local copies: insert a comment like

  // TODO(MIGRATE): LOCAL_COPY resource definitions — use models.GetDefinitions()

Contact
- If you add a new resource, update `models/resource_definitions.go`, add necessary seeding changes (see `scripts/seed_defaults`), and update this README with any special considerations.
