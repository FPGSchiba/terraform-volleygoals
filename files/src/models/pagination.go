package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
)

type Cursor struct {
	LastID        string `json:"last_id,omitempty"`
	LastCreatedAt string `json:"last_created_at,omitempty"`
}

type PaginationRequest struct {
	Limit     int
	NextToken string
	Cursor    *Cursor
}

type PaginationResponse struct {
	Items     interface{} `json:"items"`
	Count     int         `json:"count"`
	NextToken string      `json:"next_token,omitempty"`
	HasMore   bool        `json:"has_more"`
}

// EncodeCursor encodes a Cursor into an opaque next_token
func EncodeCursor(c *Cursor) (string, error) {
	if c == nil {
		return "", nil
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeCursor decodes an opaque next_token into a Cursor
func DecodeCursor(token string) (*Cursor, error) {
	if token == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ParseLimit parses limit with a sane default and max
func ParseLimit(s string, defaultLimit, maxLimit int) (int, error) {
	if s == "" {
		return defaultLimit, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, errors.New("limit must be positive")
	}
	if n > maxLimit {
		return maxLimit, nil
	}
	return n, nil
}
