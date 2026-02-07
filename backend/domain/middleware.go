package domain

import "fmt"

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

type MiddlewareInfo struct {
	ID        string       `json:"id"`
	Schema    PluginSchema `json:"schema"`
	IsDefault bool         `json:"is_default,omitempty"`
}

type Middleware interface {
	Execute(req *Request, next Handler) (*Response, error)
}

type MiddlewareFunc func(req *Request, next Handler) (*Response, error)

func (f MiddlewareFunc) Execute(req *Request, next Handler) (*Response, error) {
	return f(req, next)
}

type MiddlewareRegistry interface {
	Register(info MiddlewareInfo, factory func(config map[string]any) Middleware)
	Get(id string, config map[string]any) (Middleware, error)
	List() []MiddlewareInfo
}

type DefaultMiddlewareRegistry struct {
	plugins map[string]regEntry
}

type regEntry struct {
	factory   func(config map[string]any) Middleware
	schema    PluginSchema
	isDefault bool
}

func NewMiddlewareRegistry() MiddlewareRegistry {
	return &DefaultMiddlewareRegistry{plugins: make(map[string]regEntry)}
}

func (r *DefaultMiddlewareRegistry) Register(info MiddlewareInfo, factory func(config map[string]any) Middleware) {
	r.plugins[info.ID] = regEntry{factory: factory, schema: info.Schema, isDefault: info.IsDefault}
}

func (r *DefaultMiddlewareRegistry) Get(id string, config map[string]any) (Middleware, error) {
	entry, ok := r.plugins[id]
	if !ok {
		return nil, fmt.Errorf("middleware not found: %s", id)
	}
	return entry.factory(config), nil
}

func (r *DefaultMiddlewareRegistry) List() []MiddlewareInfo {
	var infos []MiddlewareInfo
	for id, entry := range r.plugins {
		infos = append(infos, MiddlewareInfo{
			ID:        id,
			Schema:    entry.schema,
			IsDefault: entry.isDefault,
		})
	}
	return infos
}
