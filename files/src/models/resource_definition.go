package models

type ResourceDefinition struct {
    Id                   string   `json:"id"`
    Name                 string   `json:"name"`
    Description          string   `json:"description"`
    Actions              []string `json:"actions"`
    AllowedChildResources []string `json:"allowedChildResources"`
}

