package users

import (
	"fmt"
	"strconv"
	"strings"
)

// UserFilter centralizes paging / grouping / filter options for user list calls.
// Limit uses int32 to match AWS SDK expectations in callers.
type UserFilter struct {
	Limit           int32  // page size (uses default when <= 0)
	PaginationToken string // provider pagination token
	GroupName       string // optional group name for ListUsersInGroup
	Filter          string // optional Cognito filter for ListUsers
}

var (
	limitQueryName      = "limit"
	nextTokenQueryNames = []string{"next_token", "nextToken"}
	groupQueryNames     = []string{"group_name", "groupName"}
)

// getFirst returns the first present key from keys in the provided map.
func getFirst(keys []string, q map[string]string) (string, bool) {
	for _, k := range keys {
		if v, ok := q[k]; ok {
			v = strings.TrimSpace(v)
			if v != "" {
				return v, true
			}
		}
	}
	return "", false
}

// limitOrDefault returns a sane limit for AWS calls
func (f *UserFilter) limitOrDefault(def int32) int32 {
	if f == nil {
		return def
	}
	if f.Limit <= 0 {
		return def
	}
	return f.Limit
}

// allowedAttributes maps accepted query parameter names (lower-cased) to Cognito attribute names.
// Only fields that exist on models.User are allowed: id/sub, email, name, preferred_username, and user status.
var allowedAttributes = map[string]string{
	"id":                 "sub",
	"sub":                "sub",
	"email":              "email",
	"name":               "name",
	"preferred_username": "preferred_username",
	"preferredUsername":  "preferred_username",
	"user_status":        "cognito:user_status",
	"userStatus":         "cognito:user_status",
	"status":             "status",
}

// preferredParamKeys defines a deterministic order to check query params for server-side filtering.
var preferredParamKeys = []string{
	"id",
	"sub",
	"email",
	"name",
	"preferred_username",
	"user_status",
	"status",
}

func isReservedKey(k string) bool {
	kl := strings.TrimSpace(strings.ToLower(k))
	if kl == limitQueryName {
		return true
	}
	for _, n := range nextTokenQueryNames {
		if strings.ToLower(n) == kl {
			return true
		}
	}
	for _, n := range groupQueryNames {
		if strings.ToLower(n) == kl {
			return true
		}
	}
	return false
}

// buildCognitoFilterFromParams builds a Cognito-compatible server-side filter string
// from the provided query parameters.
// It normalizes the query keys to lower-case and checks a deterministic list of attribute keys.
// If the value ends with '*' it is treated as a prefix (starts-with) match using '^='.
// Returns empty string when no suitable attribute param is present.
func buildCognitoFilterFromParams(q map[string]string) string {
	if q == nil || len(q) == 0 {
		return ""
	}

	ql := normalizeQuery(q)

	if attr, val, ok := findPreferredAttr(ql); ok {
		return buildSingleAttrFilter(attr, val)
	}
	if attr, val, ok := findAnyAllowedAttr(ql); ok {
		return buildSingleAttrFilter(attr, val)
	}
	return ""
}

func normalizeQuery(q map[string]string) map[string]string {
	ql := make(map[string]string, len(q))
	for k, v := range q {
		ql[strings.ToLower(strings.TrimSpace(k))] = strings.TrimSpace(v)
	}
	return ql
}

func findPreferredAttr(ql map[string]string) (attr, val string, ok bool) {
	for _, pk := range preferredParamKeys {
		if v, present := ql[pk]; present {
			if v == "" || isReservedKey(pk) {
				continue
			}
			attr = allowedAttributes[pk]
			if attr == "" {
				continue
			}
			return attr, v, true
		}
	}
	return "", "", false
}

func findAnyAllowedAttr(ql map[string]string) (attr, val string, ok bool) {
	for k, v := range ql {
		if isReservedKey(k) {
			continue
		}
		if v == "" {
			continue
		}
		if a, present := allowedAttributes[k]; present && a != "" {
			return a, v, true
		}
	}
	return "", "", false
}

func buildSingleAttrFilter(attr, rawVal string) string {
	val := rawVal
	op := "="
	if strings.HasSuffix(val, "*") {
		op = "^="
		val = strings.TrimSuffix(val, "*")
	}
	escaped := escapeForCognito(val)
	return fmt.Sprintf("%s %s \"%s\"", attr, op, escaped)
}

// escapeForCognito escapes double quotes and backslashes for inclusion in a Cognito filter string
func escapeForCognito(s string) string {
	// escape backslash first
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// then escape double quotes
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// UserFilterFromQuery builds a UserFilter from HTTP query parameters. Reserved keys (limit, next_token, group_name)
// are interpreted as pagination/group options. The first other parameter that maps to a searchable attribute is
// converted into the Cognito server-side filter string. Limit is parsed as int and set on the filter.
func UserFilterFromQuery(q map[string]string) (*UserFilter, error) {
	f := &UserFilter{}
	if q == nil {
		return f, nil
	}

	// limit
	if v, ok := q[limitQueryName]; ok {
		v = strings.TrimSpace(v)
		if v != "" {
			n, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			f.Limit = int32(n)
		}
	}

	// pagination token
	if v, ok := getFirst(nextTokenQueryNames, q); ok {
		f.PaginationToken = v
	}

	// group name
	if v, ok := getFirst(groupQueryNames, q); ok {
		f.GroupName = v
	}

	// build the Cognito filter from remaining params
	f.Filter = buildCognitoFilterFromParams(q)

	return f, nil
}
