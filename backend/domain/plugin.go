package domain

import (
	"pouch-ai/backend/util/registry"
)

type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeSelect  FieldType = "select"
)

type FieldRole string

const (
	FieldRoleLimit  FieldRole = "limit"
	FieldRolePeriod FieldRole = "period"
)

type FieldSchema struct {
	Type        FieldType `json:"type"`
	DisplayName string    `json:"display_name,omitempty"`
	Default     any       `json:"default,omitempty"`
	Description string    `json:"description,omitempty"`
	Options     []string  `json:"options,omitempty"`
	Role        FieldRole `json:"role,omitempty"`
}

type PluginSchema map[string]FieldSchema

type PluginConfig struct {
	ID     string         `json:"id"`
	Config map[string]any `json:"config,omitempty"`
}

type PluginInfo struct {
	ID        string       `json:"id"`
	Schema    PluginSchema `json:"schema"`
	IsDefault bool         `json:"is_default,omitempty"`
}

type MiddlewareRegistry registry.Registry[func(config map[string]any) Middleware]
type ProviderRegistry registry.Registry[Provider]

func NewMiddlewareRegistry() MiddlewareRegistry {
	return registry.NewRegistry[func(config map[string]any) Middleware]()
}

func NewProviderRegistry() ProviderRegistry {
	return registry.NewRegistry[Provider]()
}
