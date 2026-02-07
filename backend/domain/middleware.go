package domain

import "fmt"

// FieldType defines the data type of a configuration field.
type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeSelect  FieldType = "select"
)

// FieldRole defines special roles for configuration fields (e.g. for UI hints).
type FieldRole string

const (
	FieldRoleLimit  FieldRole = "limit"
	FieldRolePeriod FieldRole = "period"
)

// FieldSchema describes a single configuration field in a plugin.
type FieldSchema struct {
	Type        FieldType `json:"type"`
	DisplayName string    `json:"display_name,omitempty"`
	Default     any       `json:"default,omitempty"`
	Description string    `json:"description,omitempty"`
	Options     []string  `json:"options,omitempty"`
	Role        FieldRole `json:"role,omitempty"`
}

// PluginSchema maps configuration keys to their field schemas.
type PluginSchema map[string]FieldSchema

// MiddlewareInfo contains metadata about a registered middleware.
type MiddlewareInfo struct {
	ID        string       `json:"id"`
	Schema    PluginSchema `json:"schema"`
	IsDefault bool         `json:"is_default,omitempty"`
}

// Middleware defines the interface for request interceptors.
type Middleware interface {
	Execute(req *Request, next Handler) (*Response, error)
}

// MiddlewareFunc is a function adapter for the Middleware interface.
type MiddlewareFunc func(req *Request, next Handler) (*Response, error)

// Execute calls the underlying function.
func (f MiddlewareFunc) Execute(req *Request, next Handler) (*Response, error) {
	return f(req, next)
}

// MiddlewareRegistry manages the registration and instantiation of Middlewares.
type MiddlewareRegistry interface {
	Register(info MiddlewareInfo, factory func(config map[string]any) Middleware)
	Get(id string, config map[string]any) (Middleware, error)
	List() []MiddlewareInfo
}

// DefaultMiddlewareRegistry is the standard implementation of MiddlewareRegistry.
type DefaultMiddlewareRegistry struct {
	plugins map[string]regEntry
}

type regEntry struct {
	factory   func(config map[string]any) Middleware
	schema    PluginSchema
	isDefault bool
}

// NewMiddlewareRegistry creates a new empty DefaultMiddlewareRegistry.
func NewMiddlewareRegistry() MiddlewareRegistry {
	return &DefaultMiddlewareRegistry{plugins: make(map[string]regEntry)}
}

// Register adds a middleware factory to the registry.
func (r *DefaultMiddlewareRegistry) Register(info MiddlewareInfo, factory func(config map[string]any) Middleware) {
	r.plugins[info.ID] = regEntry{factory: factory, schema: info.Schema, isDefault: info.IsDefault}
}

// Get instantiates a middleware by ID with the provided configuration.
func (r *DefaultMiddlewareRegistry) Get(id string, config map[string]any) (Middleware, error) {
	entry, ok := r.plugins[id]
	if !ok {
		return nil, fmt.Errorf("middleware not found: %s", id)
	}
	return entry.factory(config), nil
}

// List returns metadata for all registered middlewares.
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
