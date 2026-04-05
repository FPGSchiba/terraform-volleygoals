package models

// Permission helper functions and lookup maps built from GetDefinitions().
// These helpers provide a single accessor for permission strings and simple
// lookups for resource definitions and supported actions.

import (
	"fmt"
	"strings"
)

var (
	defsByID         map[string]ResourceDefinition
	actionsByRes     map[string]map[string]bool
	permsByResAction map[string]map[string]string
)

func init() {
	defs := GetDefinitions()
	defsByID = make(map[string]ResourceDefinition, len(defs))
	actionsByRes = make(map[string]map[string]bool, len(defs))
	permsByResAction = make(map[string]map[string]string, len(defs))

	for _, d := range defs {
		defsByID[d.Id] = d
		am := make(map[string]bool)
		pm := make(map[string]string)
		for _, a := range d.Actions {
			am[a] = true
			pm[a] = fmt.Sprintf("%s:%s", d.Id, a)
		}
		actionsByRes[d.Id] = am
		permsByResAction[d.Id] = pm
	}
}

// GetPermission returns the canonical permission string for resourceID and
// action (e.g. "comments:read"). If the resource or action is unknown, it
// returns a best-effort formatted string `<resource>:<action>`.
func GetPermission(resourceID, action string) string {
	if m, ok := permsByResAction[resourceID]; ok {
		if p, ok2 := m[action]; ok2 {
			return p
		}
	}
	return fmt.Sprintf("%s:%s", resourceID, action)
}

// GetResourceByID returns the ResourceDefinition for the given id, and a
// boolean indicating whether it was found.
func GetResourceByID(resourceID string) (*ResourceDefinition, bool) {
	if d, ok := defsByID[resourceID]; ok {
		// Return a copy to avoid accidental modification of internal map value.
		dcopy := d
		return &dcopy, true
	}
	return nil, false
}

// ListResourceActions returns the list of supported actions for a resource id.
// If the resource is unknown, returns nil.
func ListResourceActions(resourceID string) []string {
	if m, ok := actionsByRes[resourceID]; ok {
		out := make([]string, 0, len(m))
		for a := range m {
			out = append(out, a)
		}
		return out
	}
	return nil
}

// ValidatePermissions checks each permission string is of the form "<resource>:<action>"
// and that the resource and action are known according to GetDefinitions(). It
// returns an error describing the first invalid permission encountered.
func ValidatePermissions(perms []string) error {
	for _, p := range perms {
		if p == "" {
			continue
		}
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid permission format: %s", p)
		}
		res := parts[0]
		act := parts[1]
		if actions, ok := actionsByRes[res]; ok {
			if !actions[act] {
				return fmt.Errorf("invalid action '%s' for resource '%s' (allowed: %v)", act, res, ListResourceActions(res))
			}
		} else {
			return fmt.Errorf("unknown resource in permission: %s", res)
		}
	}
	return nil
}

// GetAllPermissions returns the internal map of resource -> action -> permission
// string. It's safe for read-only use.
func GetAllPermissions() map[string]map[string]string {
	return permsByResAction
}
